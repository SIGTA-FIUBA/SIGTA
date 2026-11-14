package output

import (
	"context"
	"errors"
	"io"
)

// ErrNoFile also covers a digest that is not well formed, since it cannot
// name a stored file.
var ErrNoFile = errors.New("file not found")

// Files stores the files of the procedures under the SHA-256 digest of
// their content, which the store computes itself instead of trusting the
// client. The same content is stored once, and a stored file is never changed.
type Files interface {
	// Put returns the digest, "sha256:" and the hex of the hash, and the size.
	// The file is durable when Put returns, so the record that names it is
	// written after it and never points to a missing file.
	Put(ctx context.Context, content io.Reader) (digest string, size int64, err error)
	// Open fails with ErrNoFile if no file has that digest.
	Open(ctx context.Context, digest string) (io.ReadCloser, error)
}
