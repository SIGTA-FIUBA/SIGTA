package disk_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"backend/src/procedure/adapters/disk"
	"backend/src/procedure/ports/output"
)

func newDocuments(t *testing.T) (*disk.Documents, string) {
	t.Helper()
	dir := t.TempDir()
	documents, err := disk.NewDocuments(dir)
	if err != nil {
		t.Fatal(err)
	}
	return documents, dir
}

func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestPutAndOpen(t *testing.T) {
	documents, dir := newDocuments(t)
	content := []byte("%PDF-1.7 constancia sintética, año 2026 \x00\xff")
	sum := sha256.Sum256(content)
	want := "sha256:" + hex.EncodeToString(sum[:])

	for range 2 {
		digest, size, err := documents.Put(t.Context(), bytes.NewReader(content))
		if err != nil {
			t.Fatal(err)
		}
		if digest != want || size != int64(len(content)) {
			t.Errorf("Put() = %s, %d, want %s, %d", digest, size, want, len(content))
		}
	}
	if got := filesIn(t, dir); len(got) != 1 || got[0] != want[len("sha256:"):] {
		t.Errorf("files = %v, want only the one named by the digest", got)
	}

	file, err := documents.Open(t.Context(), want)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	got, err := io.ReadAll(file)
	if err != nil || !bytes.Equal(got, content) {
		t.Errorf("Open() read %q, %v, want %q", got, err, content)
	}
}

type failingReader struct{ read bool }

func (r *failingReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, errors.New("connection lost")
	}
	r.read = true
	return copy(p, "the first half"), nil
}

func TestPutLeavesNothingWhenItFails(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	for _, tc := range []struct {
		name    string
		ctx     context.Context
		content io.Reader
	}{
		{"the upload breaks", t.Context(), &failingReader{}},
		{"the request is canceled", canceled, strings.NewReader("contenido")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			documents, dir := newDocuments(t)
			if _, _, err := documents.Put(tc.ctx, tc.content); err == nil {
				t.Error("Put() = nil, want an error")
			}
			if got := filesIn(t, dir); len(got) != 0 {
				t.Errorf("files = %v, want none", got)
			}
		})
	}
}

// TestOpenOnlyByDigest plants, for each digest that is not well formed, the
// file it would open if Open went to the disk with it, so only the check of
// the digest can refuse it.
func TestOpenOnlyByDigest(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "documents")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	documents, err := disk.NewDocuments(dir)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("never stored"))
	hexSum := hex.EncodeToString(sum[:])

	for _, tc := range []struct {
		name, digest, planted string
	}{
		{"a path after the digest", "sha256:" + hexSum + "/../x", filepath.Join(dir, "x")},
		{"capital letters", "sha256:" + strings.ToUpper(hexSum), filepath.Join(dir, strings.ToUpper(hexSum))},
		{"63 characters", "sha256:" + hexSum[:63], filepath.Join(dir, hexSum[:63])},
		{"65 characters", "sha256:" + hexSum + "0", filepath.Join(dir, hexSum+"0")},
		{"the parent directory", "sha256:../outside", filepath.Join(base, "outside")},
		{"no prefix", "../outside", filepath.Join(base, "outside")},
		{"empty", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.planted != "" {
				if err := os.WriteFile(tc.planted, []byte("not a document"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			file, err := documents.Open(t.Context(), tc.digest)
			if file != nil {
				_ = file.Close()
			}
			if !errors.Is(err, output.ErrNoDocument) || file != nil {
				t.Errorf("Open(%q) = %v, %v, want no file and ErrNoDocument", tc.digest, file, err)
			}
		})
	}

	if file, err := documents.Open(t.Context(), "sha256:"+hexSum); !errors.Is(err, output.ErrNoDocument) {
		if file != nil {
			_ = file.Close()
		}
		t.Errorf("Open() of a digest that was never stored = %v, want ErrNoDocument", err)
	}
}

func TestNewDocumentsNeedsAWritableDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(dir, "missing"), file} {
		if _, err := disk.NewDocuments(path); err == nil {
			t.Errorf("NewDocuments(%s) = nil, want an error", path)
		}
	}
	if got := filesIn(t, dir); len(got) != 1 {
		t.Errorf("NewDocuments left files behind: %v", got)
	}
}
