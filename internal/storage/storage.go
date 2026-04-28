/*
Package storage provides a pluggable backend for saving and retrieving media files.

Local storage (default) keeps files on the host filesystem; S3 storage writes to
any AWS S3 or S3-compatible object store (Hetzner, MinIO, Cloudflare R2, …).
*/
package storage

import (
	"context"
	"io"
	"time"
)

// PresignExpiry is the default lifetime of a presigned read URL.
const PresignExpiry = 15 * time.Minute

// Backend is the interface that all storage implementations satisfy.
type Backend interface {
	// Put uploads a file from r to the given key.
	Put(ctx context.Context, key string, r io.Reader, size int64) error

	// Exists reports whether an object with the given key is present.
	Exists(ctx context.Context, key string) (bool, error)

	// PresignedGetURL returns a time-limited URL for reading the object at key.
	// Local backends return ("", nil); callers should fall back to local serving.
	PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
