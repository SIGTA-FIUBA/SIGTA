// Package filesystem keeps document contents in a directory, outside the
// database.
package filesystem

import (
	"backend/src/domain"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
)

// validKey keeps keys to plain names, so no key can point outside the
// directory. os.Root enforces that too; this rejects bad keys up front.
var validKey = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// DocumentStore writes each document to a file named after its key. Every
// access goes through os.Root, which refuses paths that leave the directory.
type DocumentStore struct {
	root *os.Root
}

// NewDocumentStore creates dir if needed. The caller closes the store.
func NewDocumentStore(dir string) (*DocumentStore, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &DocumentStore{root: root}, nil
}

func (s *DocumentStore) Close() error {
	return s.root.Close()
}

// Put writes to a temporary file and renames it once complete, so a reader
// never sees half a document.
func (s *DocumentStore) Put(ctx context.Context, key string, content io.Reader) (int64, error) {
	if err := checkKey(key); err != nil {
		return 0, err
	}
	tmp := ".tmp-" + rand.Text()
	f, err := s.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(f, readerWithContext{ctx, content})
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err == nil {
		err = s.root.Rename(tmp, key)
	}
	if err != nil {
		return n, errors.Join(err, ignoreMissing(s.root.Remove(tmp)))
	}
	return n, nil
}

func (s *DocumentStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	if err := checkKey(key); err != nil {
		return nil, err
	}
	f, err := s.root.Open(key)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, domain.ErrDocumentNotFound
	}
	return f, err
}

func (s *DocumentStore) Delete(_ context.Context, key string) error {
	if err := checkKey(key); err != nil {
		return err
	}
	return ignoreMissing(s.root.Remove(key))
}

func checkKey(key string) error {
	if !validKey.MatchString(key) {
		return fmt.Errorf("invalid document key %q", key)
	}
	return nil
}

func ignoreMissing(err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// readerWithContext stops a long copy when the request is canceled.
type readerWithContext struct {
	ctx context.Context
	r   io.Reader
}

func (r readerWithContext) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}
