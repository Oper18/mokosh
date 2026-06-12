package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var oncePurge sync.Once

func initPurge() {
	prg := photoprism.NewPurge(Config(), Files())
	prg.SetStorage(Storage())
	services.Purge = prg
}

// Purge returns the singleton purge worker instance.
func Purge() *photoprism.Purge {
	oncePurge.Do(initPurge)

	return services.Purge
}

// ResetPurge forces the purge worker to be re-initialised on the next call to
// Purge(). Call this alongside ResetStorage() so the purge singleton always
// picks up the current storage backend.
func ResetPurge() {
	oncePurge = sync.Once{}
	services.Purge = nil
}
