package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/storage"
)

var (
	onceStorage    sync.Once
	storageBackend storage.Backend
)

func initStorage() {
	if conf == nil {
		storageBackend = storage.NewLocal()
		return
	}

	if !conf.S3Enabled() {
		log.Debugf("storage: S3 not configured, using local storage")
		storageBackend = storage.NewLocal()
		return
	}

	aws := conf.AWS()

	backend, err := storage.NewS3(
		aws.AWSBaseEndpoint,
		aws.AWSRegion,
		aws.AWSAccessKeyID,
		aws.AWSSecretAccessKey,
		aws.S3FilesBucketName,
	)

	if err != nil {
		log.Errorf("storage: failed to initialise S3 backend (%s), falling back to local storage", err)
		storageBackend = storage.NewLocal()
		return
	}

	log.Infof("storage: using S3 bucket %q (endpoint: %q)", aws.S3FilesBucketName, aws.AWSBaseEndpoint)
	storageBackend = backend
}

// Storage returns the active file storage backend (local or S3).
func Storage() storage.Backend {
	onceStorage.Do(initStorage)
	return storageBackend
}

// ResetStorage forces the storage backend to be re-initialised on the next
// call to Storage(). Call this after SetConfig() so S3 settings take effect.
func ResetStorage() {
	onceStorage = sync.Once{}
	storageBackend = nil
}
