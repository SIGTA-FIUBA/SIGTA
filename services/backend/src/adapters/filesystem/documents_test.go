package filesystem_test

import (
	"backend/src/adapters/filesystem"
	"backend/src/domain"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newStore(t *testing.T) (*filesystem.DocumentStore, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "documents")
	store, err := filesystem.NewDocumentStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, dir
}

func TestPutOpenAndDelete(t *testing.T) {
	store, dir := newStore(t)

	n, err := store.Put(t.Context(), "doc1", strings.NewReader("contenido"))
	if err != nil || n != int64(len("contenido")) {
		t.Fatalf("Put = %d, %v", n, err)
	}

	f, err := store.Open(t.Context(), "doc1")
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(f)
	_ = f.Close()
	if err != nil || string(got) != "contenido" {
		t.Fatalf("content = %q, %v", got, err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "doc1" {
		t.Errorf("directory holds %v, want only doc1 and no temporary files", entries)
	}

	if err := store.Delete(t.Context(), "doc1"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Open(t.Context(), "doc1"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("Open after Delete: err = %v, want %v", err, domain.ErrDocumentNotFound)
	}
	if err := store.Delete(t.Context(), "doc1"); err != nil {
		t.Errorf("deleting twice: %v", err)
	}
}

func TestOpenMissingDocument(t *testing.T) {
	store, _ := newStore(t)

	if _, err := store.Open(t.Context(), "missing"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Fatalf("err = %v, want %v", err, domain.ErrDocumentNotFound)
	}
}

func TestKeysCannotLeaveTheDirectory(t *testing.T) {
	store, dir := newStore(t)
	outside := filepath.Join(filepath.Dir(dir), "outside")
	if err := os.WriteFile(outside, []byte("secreto"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"../outside", "/etc/passwd", "a/b", "", "."} {
		if _, err := store.Put(t.Context(), key, strings.NewReader("x")); err == nil {
			t.Errorf("Put(%q) succeeded", key)
		}
		if _, err := store.Open(t.Context(), key); err == nil {
			t.Errorf("Open(%q) succeeded", key)
		}
		if err := store.Delete(t.Context(), key); err == nil {
			t.Errorf("Delete(%q) succeeded", key)
		}
	}
	if _, err := os.Stat(outside); err != nil {
		t.Errorf("file outside the directory: %v", err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }

func TestFailedPutLeavesNothingBehind(t *testing.T) {
	store, dir := newStore(t)

	if _, err := store.Put(t.Context(), "doc1", io.MultiReader(strings.NewReader("parte"), failingReader{})); err == nil {
		t.Fatal("Put succeeded with a failing reader")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("directory holds %v after a failed Put", entries)
	}
}

func TestPutStopsWhenTheRequestIsCanceled(t *testing.T) {
	store, _ := newStore(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := store.Put(ctx, "doc1", strings.NewReader("contenido")); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want %v", err, context.Canceled)
	}
	if _, err := store.Open(t.Context(), "doc1"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("a canceled Put left a document: %v", err)
	}
}
