package storage

import (
	"context"
	"io"
	"time"
)

// Local is a no-op Backend for the local filesystem.
// File writes are handled by existing copy/move code; this backend only
// satisfies the interface so callers can remain storage-agnostic.
type Local struct{}

// NewLocal creates a local storage backend.
func NewLocal() *Local {
	return &Local{}
}

func (l *Local) Put(_ context.Context, _ string, _ io.Reader, _ int64) error {
	return nil
}

func (l *Local) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// PresignedGetURL always returns an empty string for local storage,
// signalling callers to serve the file directly from disk.
func (l *Local) PresignedGetURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}
