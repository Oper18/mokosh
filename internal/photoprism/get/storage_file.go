package get

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/storage"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// remoteEligible reports whether a file is a candidate for remote storage, i.e.
// it lives under the originals root and has a non-empty relative name.
func remoteEligible(f *entity.File) bool {
	if f == nil || f.FileName == "" || f.FileRoot != entity.RootOriginals {
		return false
	}

	b := Storage()

	if b == nil {
		return false
	} else if _, isLocal := b.(*storage.Local); isLocal {
		return false
	}

	return true
}

// RemoteMayHold reports whether remote storage either holds the file or could
// not be reached. In both cases the file must not be flagged as missing or have
// its photo trashed. It returns false for the local backend and for objects the
// store positively reports as absent, preserving the original behaviour.
func RemoteMayHold(f *entity.File) bool {
	if !remoteEligible(f) {
		return false
	}

	exists, err := Storage().Exists(context.Background(), f.FileName)

	if err != nil {
		// Unreachable store: keep the file rather than risk destroying data.
		log.Warnf("storage: could not verify %s in remote storage, keeping it (%s)", clean.Log(f.FileName), err)
		return true
	}

	return exists
}

// ResolveLocalFile returns a path to the file on the local filesystem. When the
// original is not present locally but exists in remote storage (e.g. S3), it is
// fetched into the local originals cache and that path is returned. The bool
// result reports whether the file was fetched from remote storage.
func ResolveLocalFile(f *entity.File) (localName string, fetched bool, err error) {
	fileName := photoprism.FileName(f.FileRoot, f.FileName)

	// Prefer the local copy if it already exists.
	if resolved, resolveErr := fs.Resolve(fileName); resolveErr == nil {
		return resolved, false, nil
	}

	if !remoteEligible(f) {
		return "", false, os.ErrNotExist
	}

	// Stream the object from remote storage into the originals cache atomically.
	rc, getErr := Storage().Get(context.Background(), f.FileName)
	if getErr != nil {
		return "", false, getErr
	}
	defer rc.Close()

	if mkErr := os.MkdirAll(filepath.Dir(fileName), fs.ModeDir); mkErr != nil {
		return "", false, mkErr
	}

	tmp, tmpErr := os.CreateTemp(filepath.Dir(fileName), ".storage-*")
	if tmpErr != nil {
		return "", false, tmpErr
	}
	tmpName := tmp.Name()

	if _, copyErr := io.Copy(tmp, rc); copyErr != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return "", false, copyErr
	}

	if closeErr := tmp.Close(); closeErr != nil {
		_ = os.Remove(tmpName)
		return "", false, closeErr
	}

	if renErr := os.Rename(tmpName, fileName); renErr != nil {
		_ = os.Remove(tmpName)
		return "", false, renErr
	}

	log.Debugf("storage: fetched %s from remote storage", clean.Log(f.FileName))

	if _, resolveErr := fs.Resolve(fileName); resolveErr != nil {
		return "", true, fmt.Errorf("storage: fetched %s but it is not readable (%s)", clean.Log(f.FileName), resolveErr)
	}

	return fileName, true, nil
}
