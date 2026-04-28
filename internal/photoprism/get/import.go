package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceImport sync.Once

func initImport() {
	imp := photoprism.NewImport(Config(), Index(), Convert())
	imp.SetStorage(Storage())
	services.Import = imp
}

// Import returns the singleton import service instance.
func Import() *photoprism.Import {
	onceImport.Do(initImport)

	return services.Import
}

// ResetImport forces the import service to be re-initialised on the next call
// to Import(). Call this alongside ResetStorage() so the import singleton
// always picks up the current storage backend.
func ResetImport() {
	onceImport = sync.Once{}
	services.Import = nil
}
