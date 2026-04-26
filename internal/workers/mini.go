package workers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	MiniImageWidth  = 512
	MiniVideoHeight = 360
	MiniSuffix      = "_mini"
)

// Mini represents a worker that generates degraded copies of media files for
// low-permission user shares (PermComment or below).
type Mini struct {
	conf *config.Config
}

// NewMini returns a new Mini worker.
func NewMini(conf *config.Config) *Mini {
	return &Mini{conf: conf}
}

// ProcessShare generates mini versions for all files associated with the share.
// shareUID may be a photo UID or album UID. ownerName is used as the watermark text.
func (w *Mini) ProcessShare(shareUID, ownerName string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("mini: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if shareUID == "" {
		return fmt.Errorf("mini: share uid is empty")
	}

	switch {
	case rnd.IsUID(shareUID, entity.PhotoUID):
		return w.processPhoto(shareUID, ownerName)
	case rnd.IsUID(shareUID, entity.AlbumUID):
		return w.processAlbum(shareUID, ownerName)
	default:
		return fmt.Errorf("mini: unsupported share uid %s", shareUID)
	}
}

// processPhoto generates mini files for the primary image and video of a photo.
func (w *Mini) processPhoto(photoUID, ownerName string) error {
	// Process primary image.
	if imgFile, err := query.FileByPhotoUID(photoUID); err == nil {
		if genErr := w.generateMiniImage(imgFile, ownerName); genErr != nil {
			log.Warnf("mini: %s (image)", genErr)
		}
	}

	// Process video if present.
	if vidFile, err := query.VideoByPhotoUID(photoUID); err == nil && vidFile.FileVideo {
		if genErr := w.generateMiniVideo(vidFile, ownerName); genErr != nil {
			log.Warnf("mini: %s (video)", genErr)
		}
	}

	return nil
}

// processAlbum generates mini files for every photo in the album.
func (w *Mini) processAlbum(albumUID, ownerName string) error {
	var photoUIDs []string

	if err := entity.Db().
		Table("photos_albums").
		Where("album_uid = ? AND hidden = 0 AND missing = 0", albumUID).
		Pluck("photo_uid", &photoUIDs).Error; err != nil {
		return fmt.Errorf("mini: failed to list album photos: %w", err)
	}

	for _, photoUID := range photoUIDs {
		if err := w.processPhoto(photoUID, ownerName); err != nil {
			log.Warnf("mini: %s (photo %s)", err, photoUID)
		}
	}

	return nil
}

// generateMiniImage creates a 512px-wide watermarked JPEG copy of the source image.
func (w *Mini) generateMiniImage(f *entity.File, watermark string) error {
	srcPath := photoprism.FileName(f.FileRoot, f.FileName)

	if !fs.FileExistsNotEmpty(srcPath) {
		return fmt.Errorf("source image not found: %s", f.FileName)
	}

	destPath, err := w.miniPath(f.FileHash, ".jpg")
	if err != nil {
		return err
	}

	if fs.FileExistsNotEmpty(destPath) {
		return nil
	}

	// scale=512:-1 preserves aspect ratio, drawtext adds watermark.
	vf := fmt.Sprintf(
		"scale=%d:-1,drawtext=text='%s':fontsize=24:fontcolor=white:x=10:y=10:shadowcolor=black:shadowx=1:shadowy=1",
		MiniImageWidth, escapeFFmpegText(watermark),
	)

	// #nosec G204 -- arguments built from validated config and DB-sourced paths.
	cmd := exec.Command(
		w.conf.FFmpegBin(),
		"-hide_banner", "-y",
		"-i", srcPath,
		"-vf", vf,
		"-q:v", "5",
		destPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg image mini failed for %s: %w", f.FileName, err)
	}

	log.Infof("mini: created image %s", filepath.Base(destPath))

	return nil
}

// generateMiniVideo creates a 360p watermarked MP4 copy of the source video.
func (w *Mini) generateMiniVideo(f *entity.File, watermark string) error {
	srcPath := photoprism.FileName(f.FileRoot, f.FileName)

	if !fs.FileExistsNotEmpty(srcPath) {
		return fmt.Errorf("source video not found: %s", f.FileName)
	}

	destPath, err := w.miniPath(f.FileHash, ".mp4")
	if err != nil {
		return err
	}

	if fs.FileExistsNotEmpty(destPath) {
		return nil
	}

	// scale=-2:360 downscales to 360p, drawtext adds watermark.
	vf := fmt.Sprintf(
		"scale=-2:%d,drawtext=text='%s':fontsize=24:fontcolor=white:x=10:y=10:shadowcolor=black:shadowx=1:shadowy=1",
		MiniVideoHeight, escapeFFmpegText(watermark),
	)

	// #nosec G204 -- arguments built from validated config and DB-sourced paths.
	cmd := exec.Command(
		w.conf.FFmpegBin(),
		"-hide_banner", "-y",
		"-i", srcPath,
		"-vf", vf,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "28",
		"-c:a", "aac",
		"-movflags", "+faststart",
		"-f", "mp4",
		destPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg video mini failed for %s: %w", f.FileName, err)
	}

	log.Infof("mini: created video %s", filepath.Base(destPath))

	return nil
}

// miniPath returns the storage path for a mini file, creating directories as needed.
// The filename format is: <storageRoot>/mini/<h[0]>/<h[1]>/<hash>_mini<ext>
func (w *Mini) miniPath(hash, ext string) (string, error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("mini: file hash too short")
	}

	dir := filepath.Join(w.conf.StoragePath(), "mini", hash[0:1], hash[1:2])

	if err := fs.MkdirAll(dir); err != nil {
		return "", fmt.Errorf("mini: failed to create directory %s: %w", dir, err)
	}

	return filepath.Join(dir, hash+MiniSuffix+ext), nil
}

// MiniImagePath returns the expected mini image path for a given file hash.
// Returns an empty string if the file does not exist yet.
func MiniImagePath(conf *config.Config, hash string) string {
	if len(hash) < 4 {
		return ""
	}

	p := filepath.Join(conf.StoragePath(), "mini", hash[0:1], hash[1:2], hash+MiniSuffix+".jpg")

	if fs.FileExistsNotEmpty(p) {
		return p
	}

	return ""
}

// MiniVideoPath returns the expected mini video path for a given file hash.
// Returns an empty string if the file does not exist yet.
func MiniVideoPath(conf *config.Config, hash string) string {
	if len(hash) < 4 {
		return ""
	}

	p := filepath.Join(conf.StoragePath(), "mini", hash[0:1], hash[1:2], hash+MiniSuffix+".mp4")

	if fs.FileExistsNotEmpty(p) {
		return p
	}

	return ""
}

// escapeFFmpegText escapes characters that are special in FFmpeg drawtext filter values.
func escapeFFmpegText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, `:`, `\:`)
	return s
}
