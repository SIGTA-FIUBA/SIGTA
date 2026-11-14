// Package disk stores the documents of the procedures as files in a directory.
// Nothing uses it yet: it gets wired to the server along with the upload of
// documents.
package disk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"backend/src/procedure/ports/output"
)

const (
	digestPrefix = "sha256:"
	tempPrefix   = ".tmp-"

	// staleTemp is how old a temporary file has to be for NewDocuments to
	// take it for what a crash left, and not for an upload in progress.
	staleTemp = time.Hour
)

// validDigest also keeps a digest from naming a path outside the directory.
var validDigest = regexp.MustCompile("^" + regexp.QuoteMeta(digestPrefix) + "[0-9a-f]{64}$")

var _ output.Documents = (*Documents)(nil)

// Documents names each file with the hex of its digest. A file is written
// under a temporary name, synced and then renamed, so a crash never leaves a
// partial file under a digest.
type Documents struct {
	dir string
}

// NewDocuments fails if dir is not a directory the server can write to, so a
// missing volume is noticed at startup and not on the first upload. It also
// removes the temporary files that a crash left.
func NewDocuments(dir string) (*Documents, error) {
	probe, err := os.CreateTemp(dir, ".probe-")
	if err != nil {
		return nil, fmt.Errorf("documents directory %s: %w", dir, err)
	}
	if err := errors.Join(probe.Close(), os.Remove(probe.Name()), removeStaleTemps(dir)); err != nil {
		return nil, fmt.Errorf("documents directory %s: %w", dir, err)
	}
	return &Documents{dir: dir}, nil
}

func removeStaleTemps(dir string) error {
	temps, err := filepath.Glob(filepath.Join(dir, tempPrefix+"*"))
	if err != nil {
		return err
	}
	var errs []error
	for _, temp := range temps {
		info, err := os.Stat(temp)
		if err == nil && time.Since(info.ModTime()) > staleTemp {
			err = os.Remove(temp)
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (d *Documents) Put(ctx context.Context, content io.Reader) (digest string, size int64, err error) {
	tmp, err := os.CreateTemp(d.dir, tempPrefix)
	if err != nil {
		return "", 0, fmt.Errorf("put document: %w", err)
	}
	renamed := false
	defer func() {
		if !renamed {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()

	hash := sha256.New()
	size, err = io.Copy(io.MultiWriter(tmp, hash), contextReader{ctx, content})
	if err != nil {
		return "", 0, fmt.Errorf("put document: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return "", 0, fmt.Errorf("put document: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", 0, fmt.Errorf("put document: %w", err)
	}
	name := hex.EncodeToString(hash.Sum(nil))
	// Renaming over an existing file is harmless: it has the same content.
	if err := os.Rename(tmp.Name(), filepath.Join(d.dir, name)); err != nil {
		return "", 0, fmt.Errorf("put document: %w", err)
	}
	renamed = true
	// The rename is durable only once the directory is synced.
	if err := syncDir(d.dir); err != nil {
		return "", 0, fmt.Errorf("put document: %w", err)
	}
	return digestPrefix + name, size, nil
}

func (d *Documents) Open(_ context.Context, digest string) (io.ReadCloser, error) {
	if !validDigest.MatchString(digest) {
		return nil, fmt.Errorf("%w: invalid digest %q", output.ErrNoDocument, digest)
	}
	file, err := os.Open(filepath.Join(d.dir, digest[len(digestPrefix):]))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", output.ErrNoDocument, digest)
	}
	if err != nil {
		return nil, fmt.Errorf("open document %s: %w", digest, err)
	}
	return file, nil
}

// contextReader stops reading once ctx is done, so a canceled request does
// not keep writing its upload.
type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

func syncDir(dir string) error {
	f, err := os.Open(dir) //nolint:gosec // G304: the directory comes from configuration.
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}
