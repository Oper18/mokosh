package photoprism

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/storage"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Purge represents a worker that removes missing files from search results.
type Purge struct {
	conf    *config.Config
	files   *Files
	storage storage.Backend
}

// NewPurge returns a new purge worker.
func NewPurge(conf *config.Config, files *Files) *Purge {
	instance := &Purge{
		conf:    conf,
		files:   files,
		storage: storage.NewLocal(),
	}

	return instance
}

// SetStorage sets the remote storage backend consulted before a file is flagged
// as missing. Objects that still exist in remote storage (e.g. S3) are then kept
// even when absent from the local cache, so they are never purged or trashed.
func (w *Purge) SetStorage(b storage.Backend) {
	if b != nil {
		w.storage = b
	}
}

// remoteState describes what the configured remote storage backend knows about
// a file.
type remoteState int

const (
	remoteNotApplicable remoteState = iota // local backend or file not stored remotely — no remote opinion
	remotePresent                          // object confirmed present in remote storage
	remoteAbsent                           // object confirmed absent (HTTP 404)
	remoteUnreachable                      // remote configured but the check failed (transient)
)

// remoteFileState reports what remote storage knows about the originals file
// identified by its root and relative name. It returns remoteNotApplicable for
// the local backend and for files that are not stored remotely, and
// remoteUnreachable when the backend cannot be reached — so a transient outage
// is never treated as a deletion.
func (w *Purge) remoteFileState(fileRoot, fileName string) remoteState {
	if w.storage == nil {
		return remoteNotApplicable
	} else if _, isLocal := w.storage.(*storage.Local); isLocal {
		return remoteNotApplicable
	} else if fileRoot != entity.RootOriginals || fileName == "" {
		return remoteNotApplicable
	}

	switch exists, err := w.storage.Exists(context.Background(), fileName); {
	case err != nil:
		log.Warnf("purge: could not verify %s in remote storage, leaving it unchanged (%s)", clean.Log(fileName), err)
		return remoteUnreachable
	case exists:
		return remotePresent
	default:
		return remoteAbsent
	}
}

// remoteMayHold reports whether remote storage either holds the file or could
// not be reached. In both cases the file must not be flagged as missing. It
// returns false for the local backend and for confirmed-absent objects, which
// preserves the original local-only purge behaviour.
func (w *Purge) remoteMayHold(fileRoot, fileName string) bool {
	switch w.remoteFileState(fileRoot, fileName) {
	case remotePresent, remoteUnreachable:
		return true
	default:
		return false
	}
}

// Start removes missing files from search results.
func (w *Purge) Start(opt PurgeOptions) (purgedFiles map[string]bool, purgedPhotos map[string]bool, updated int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("purge: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	originalsPath := w.conf.OriginalsPath()

	// Check if originals folder is empty.
	if fs.DirIsEmpty(originalsPath) {
		return purgedFiles, purgedPhotos, 0, err
	}

	var ignore fs.Done

	if opt.Ignore != nil {
		ignore = opt.Ignore
	} else {
		ignore = make(fs.Done)
	}

	purgedFiles = make(map[string]bool)
	purgedPhotos = make(map[string]bool)

	if err = mutex.IndexWorker.Start(); err != nil {
		log.Warnf("purge: %s (start)", err.Error())
		return purgedFiles, purgedPhotos, 0, err
	}

	defer mutex.IndexWorker.Stop()

	// Count updates.
	updatedFiles := 0
	updatedDuplicates := 0
	updatedPhotos := 0

	// Total number of updates.
	updates := func() int {
		return updatedFiles + updatedDuplicates + updatedPhotos
	}

	// Check files.
	startFiles := time.Now()
	limit := 10000
	offset := 0
	for {
		files, err := query.Files(limit, offset, opt.Path, true)

		if err != nil {
			return purgedFiles, purgedPhotos, updates(), err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			if mutex.IndexWorker.Canceled() {
				return purgedFiles, purgedPhotos, updates(), errors.New("purge canceled")
			}

			fileName := FileName(file.FileRoot, file.FileName)

			if ignore[fileName].Exists() || purgedFiles[fileName] {
				continue
			}

			if file.FileMissing {
				if fs.FileExists(fileName) || w.remoteFileState(file.FileRoot, file.FileName) == remotePresent {
					if opt.Dry {
						log.Infof("purge: found %s", clean.Log(file.FileName))
						continue
					}

					updatedFiles++

					if err := file.Found(); err != nil {
						log.Errorf("purge: %s", err)
					} else {
						log.Infof("purge: found %s", clean.Log(file.FileName))
					}
				}
			} else if !fs.FileExists(fileName) && !w.remoteMayHold(file.FileRoot, file.FileName) {
				if opt.Dry {
					purgedFiles[fileName] = true
					log.Infof("purge: file %s would be flagged as missing", clean.Log(file.FileName))
					continue
				}

				updatedFiles++

				wasPrimary := file.FilePrimary

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
					continue
				}

				w.files.Remove(file.FileName, file.FileRoot)
				purgedFiles[fileName] = true
				log.Infof("purge: flagged file %s as missing", clean.Log(file.FileName))

				if !wasPrimary {
					continue
				}

				if err := query.SetPhotoPrimary(file.PhotoUID, ""); err != nil {
					log.Infof("purge: %s", err)
				}
			}
		}

		if mutex.IndexWorker.Canceled() {
			return purgedFiles, purgedPhotos, updates(), errors.New("purge canceled")
		}

		offset += limit
	}

	log.Debugf("purge: updated %s [%s]", english.Plural(updatedFiles, "file", "files"), time.Since(startFiles))

	// Check duplicates.
	startDuplicates := time.Now()
	limit = 10000
	offset = 0
	for {
		files, err := query.Duplicates(limit, offset, opt.Path)

		if err != nil {
			return purgedFiles, purgedPhotos, updates(), err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			if mutex.IndexWorker.Canceled() {
				return purgedFiles, purgedPhotos, updates(), errors.New("purge canceled")
			}

			fileName := FileName(file.FileRoot, file.FileName)

			if ignore[fileName].Exists() || purgedFiles[fileName] {
				continue
			}

			if !fs.FileExists(fileName) && !w.remoteMayHold(file.FileRoot, file.FileName) {
				if opt.Dry {
					purgedFiles[fileName] = true
					log.Infof("purge: duplicate %s would be removed from index", clean.Log(file.FileName))
					continue
				}

				updatedDuplicates++

				if err := file.Purge(); err != nil {
					log.Errorf("purge: %s", err)
				} else {
					w.files.Remove(file.FileName, file.FileRoot)
					purgedFiles[fileName] = true
					log.Infof("purge: removed duplicate %s from index", clean.Log(file.FileName))
				}
			}
		}

		if mutex.IndexWorker.Canceled() {
			return purgedFiles, purgedPhotos, updates(), errors.New("purge canceled")
		}

		offset += limit
	}

	log.Debugf("purge: updated %s [%s]", english.Plural(updatedDuplicates, "duplicate", "duplicates"), time.Since(startDuplicates))

	// Check photos.
	startPhotos := time.Now()
	limit = 10000
	offset = 0
	for {
		photos, err := query.MissingPhotos(limit, offset)

		if err != nil {
			return purgedFiles, purgedPhotos, updates(), err
		}

		if len(photos) == 0 {
			break
		}

		for _, photo := range photos {
			if mutex.IndexWorker.Canceled() {
				return purgedFiles, purgedPhotos, updates(), errors.New("purge canceled")
			}

			if purgedPhotos[photo.PhotoUID] {
				continue
			}

			if opt.Dry {
				purgedPhotos[photo.PhotoUID] = true
				log.Infof("purge: %s would be removed", photo.String())
				continue
			}

			updatedPhotos++

			if files, err := photo.Delete(opt.Hard); err != nil {
				log.Errorf("purge: %s (delete photo)", err)
			} else {
				purgedPhotos[photo.PhotoUID] = true

				if opt.Hard {
					log.Infof("purge: permanently removed %s", photo.String())
				} else {
					log.Infof("purge: flagged photo %s as deleted", photo.String())
				}

				// Remove files from lookup table.
				for _, file := range files {
					w.files.Remove(file.FileName, file.FileRoot)
				}
			}
		}

		if mutex.IndexWorker.Canceled() {
			return purgedFiles, purgedPhotos, updates(), errors.New("purge canceled")
		}

		offset += limit
	}

	log.Debugf("purge: updated %s [%s]", english.Plural(updatedPhotos, "photo", "photos"), time.Since(startPhotos))

	// Skip the index update if there are no changes.
	if !opt.Force && updates() == 0 {
		return purgedFiles, purgedPhotos, updates(), nil
	}

	startIndex := time.Now()

	if err = query.FixPrimaries(); err != nil {
		log.Errorf("index: %s (update primary files)", err)
	}

	// Set photo quality scores to -1 if files are missing.
	if err = query.FlagHiddenPhotos(); err != nil {
		return purgedFiles, purgedPhotos, updates(), err
	}

	// Remove orphan index entries.
	if opt.Dry {
		if files, err := query.OrphanFiles(); err != nil {
			log.Errorf("index: %s (find orphan files)", err)
		} else if l := len(files); l > 0 {
			log.Infof("index: found %s", english.Plural(l, "orphan file", "orphan files"))
		} else {
			log.Infof("index: found no orphan files")
		}
	} else {
		if err = query.PurgeOrphans(); err != nil {
			log.Errorf("index: %s (purge orphans)", err)
		}

		// Regenerate search index columns.
		entity.File{}.RegenerateIndex()
	}

	// Hide missing album contents.
	if err = query.UpdateMissingAlbumEntries(); err != nil {
		log.Errorf("index: %s (update album entries)", err)
	}

	// Remove unused entries from the places table.
	if err = query.PurgePlaces(); err != nil {
		log.Errorf("index: %s (purge places)", err)
	}

	// Update precalculated photo and file counts.
	if err = entity.UpdateCounts(); err != nil {
		log.Warnf("index: %s (update counts)", err)
	}

	// Update album, subject, and label cover thumbs.
	if err = query.UpdateCovers(); err != nil {
		log.Warnf("index: %s (update covers)", err)
	}

	log.Debugf("purge: updated index [%s]", time.Since(startIndex))

	return purgedFiles, purgedPhotos, updates(), nil
}

// Cancel stops the current operation.
func (w *Purge) Cancel() {
	mutex.IndexWorker.Cancel()
}
