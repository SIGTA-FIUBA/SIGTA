package domain

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readJSON rejects unknown keys, so that a misspelled key fails instead of
// being dropped.
func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	raw, err := os.ReadFile(path) //nolint:gosec // The test chooses the path, not a user.
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

// TestDrafts validates drafts of procedure types kept outside the repository,
// since they describe processes the Secretaría has not confirmed. It reads
// every JSON file in the folder PROCEDURE_DRAFTS names, with the areas they
// use in areas.json.
func TestDrafts(t *testing.T) {
	dir := os.Getenv("PROCEDURE_DRAFTS")
	if dir == "" {
		t.Skip("PROCEDURE_DRAFTS is not set")
	}
	var areas []Area
	readJSON(t, filepath.Join(dir, "areas.json"), &areas)
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) < 2 {
		t.Fatalf("%s has no drafts besides areas.json", dir)
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".json")
		if name == "areas" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			var d Definition
			readJSON(t, path, &d)
			if err := d.Validate(areas); err != nil {
				t.Errorf("Validate() = %v", err)
			}
		})
	}
}
