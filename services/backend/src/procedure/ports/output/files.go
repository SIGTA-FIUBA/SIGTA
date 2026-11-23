package output

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrNoFile also covers a digest that is not well formed, since it cannot
// name a stored file.
var ErrNoFile = errors.New("file not found")

// ErrCorruptFile is a stored file whose content no longer has its digest.
var ErrCorruptFile = errors.New("stored file does not match its digest")

// Files stores the files of the procedures under the SHA-256 digest of
// their content, which the store computes itself instead of trusting the
// client. The same content is stored once, and a stored file is never changed:
// it is only removed once no document names it.
type Files interface {
	// Put returns the digest, "sha256:" and the hex of the hash, and the size.
	// The file is durable when Put returns, so the record that names it is
	// written after it and never points to a missing file.
	Put(ctx context.Context, content io.Reader) (digest string, size int64, err error)
	// Open fails with ErrNoFile if no file has that digest, and with
	// ErrCorruptFile if the content of the file no longer has it.
	Open(ctx context.Context, digest string) (io.ReadCloser, error)
	// Digests returns the digests of the files stored before storedBefore,
	// in no order. A file stored again counts from the last time.
	Digests(ctx context.Context, storedBefore time.Time) ([]string, error)
	// Remove deletes the file with that digest, and does nothing if there is
	// none.
	Remove(ctx context.Context, digest string) error
}
