// Package disk stores the files of the procedures in a directory.
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

	// staleTemp is how old a temporary file has to be for NewFiles to
	// take it for what a crash left, and not for an upload in progress.
	staleTemp = time.Hour
)

// validDigest also keeps a digest from naming a path outside the directory.
var validDigest = regexp.MustCompile("^" + regexp.QuoteMeta(digestPrefix) + "[0-9a-f]{64}$")

var _ output.Files = (*Files)(nil)

// Files names each file with the hex of its digest. A file is written
// under a temporary name, synced and then renamed, so a crash never leaves a
// partial file under a digest.
type Files struct {
	dir string
}

// NewFiles fails if dir is not a directory the server can write to, so a
// missing volume is noticed at startup and not on the first upload. It also
// removes the temporary files that a crash left.
func NewFiles(dir string) (*Files, error) {
	probe, err := os.CreateTemp(dir, ".probe-")
	if err != nil {
		return nil, fmt.Errorf("documents directory %s: %w", dir, err)
	}
	if err := errors.Join(probe.Close(), os.Remove(probe.Name()), removeStaleTemps(dir)); err != nil {
		return nil, fmt.Errorf("documents directory %s: %w", dir, err)
	}
	return &Files{dir: dir}, nil
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

func (f *Files) Put(ctx context.Context, content io.Reader) (digest string, size int64, err error) {
	tmp, err := os.CreateTemp(f.dir, tempPrefix)
	if err != nil {
		return "", 0, fmt.Errorf("put file: %w", err)
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
		return "", 0, fmt.Errorf("put file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return "", 0, fmt.Errorf("put file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", 0, fmt.Errorf("put file: %w", err)
	}
	name := hex.EncodeToString(hash.Sum(nil))
	// Renaming over an existing file is harmless: it has the same content.
	if err := os.Rename(tmp.Name(), filepath.Join(f.dir, name)); err != nil {
		return "", 0, fmt.Errorf("put file: %w", err)
	}
	renamed = true
	// The rename is durable only once the directory is synced.
	if err := syncDir(f.dir); err != nil {
		return "", 0, fmt.Errorf("put file: %w", err)
	}
	return digestPrefix + name, size, nil
}

func (f *Files) Open(_ context.Context, digest string) (io.ReadCloser, error) {
	if !validDigest.MatchString(digest) {
		return nil, fmt.Errorf("%w: invalid digest %q", output.ErrNoFile, digest)
	}
	file, err := os.Open(filepath.Join(f.dir, digest[len(digestPrefix):]))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", output.ErrNoFile, digest)
	}
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", digest, err)
	}
	return file, nil
}

func (f *Files) Digests(_ context.Context) ([]string, error) {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	var digests []string
	for _, entry := range entries {
		// The temporary files and anything else in the directory are not
		// named by a digest.
		if digest := digestPrefix + entry.Name(); entry.Type().IsRegular() && validDigest.MatchString(digest) {
			digests = append(digests, digest)
		}
	}
	return digests, nil
}

func (f *Files) Remove(_ context.Context, digest string) error {
	if !validDigest.MatchString(digest) {
		return fmt.Errorf("%w: invalid digest %q", output.ErrNoFile, digest)
	}
	err := os.Remove(filepath.Join(f.dir, digest[len(digestPrefix):]))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove file %s: %w", digest, err)
	}
	// Like the rename of Put, the removal is durable once the directory is
	// synced.
	if err := syncDir(f.dir); err != nil {
		return fmt.Errorf("remove file %s: %w", digest, err)
	}
	return nil
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
