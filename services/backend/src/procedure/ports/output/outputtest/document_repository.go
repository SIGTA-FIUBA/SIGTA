package outputtest

import (
	"errors"
	"slices"
	"testing"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

// TestDocumentRepository runs the contract of output.DocumentRepository.
// newRepository returns an empty repository each time it is called.
func TestDocumentRepository(t *testing.T, newRepository func(t *testing.T) output.DocumentRepository) {
	t.Run("in the order they were added", func(t *testing.T) { inOrder(t, newRepository(t)) })
	t.Run("only in their procedure", func(t *testing.T) { onlyInTheirProcedure(t, newRepository(t)) })
	t.Run("digests of every procedure", func(t *testing.T) { digestsOfEveryProcedure(t, newRepository(t)) })
}

func document(id, procedureID, requirement string) domain.Document {
	return domain.Document{
		ID:          id,
		Procedure:   procedureID,
		Requirement: requirement,
		Step:        "documents",
		FileName:    id + ".pdf",
		MediaType:   "application/pdf",
		Size:        100,
		Digest:      "sha256:" + id,
		UploadedBy:  "titular@fi.uba.ar",
		UploadedAt:  time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC),
	}
}

func addAll(t *testing.T, r output.DocumentRepository, documents ...domain.Document) {
	t.Helper()
	for _, d := range documents {
		if err := r.Add(t.Context(), d); err != nil {
			t.Fatalf("Add(%s) = %v", d.ID, err)
		}
	}
}

func inOrder(t *testing.T, r output.DocumentRepository) {
	added := []domain.Document{
		document("D-3", "P-1", "statement"),
		document("D-1", "P-1", "annex"),
		document("D-2", "P-1", "statement"),
	}
	addAll(t, r, added...)
	got, err := r.List(t.Context(), "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, added) {
		t.Errorf("List = %+v, want %+v", got, added)
	}
}

func onlyInTheirProcedure(t *testing.T, r output.DocumentRepository) {
	addAll(t, r, document("D-1", "P-1", "statement"), document("D-2", "P-2", "statement"))
	got, err := r.List(t.Context(), "P-2")
	if err != nil || len(got) != 1 || got[0].ID != "D-2" {
		t.Errorf("List(P-2) = %+v, %v; want only D-2", got, err)
	}
	if empty, err := r.List(t.Context(), "P-3"); err != nil || len(empty) != 0 {
		t.Errorf("List(P-3) = %+v, %v; want none", empty, err)
	}
	if d, err := r.Get(t.Context(), "P-1", "D-1"); err != nil || d != document("D-1", "P-1", "statement") {
		t.Errorf("Get(P-1, D-1) = %+v, %v", d, err)
	}
	if _, err := r.Get(t.Context(), "P-2", "D-1"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("Get(P-2, D-1) = %v, want ErrDocumentNotFound", err)
	}
}

func digestsOfEveryProcedure(t *testing.T, r output.DocumentRepository) {
	again := document("D-3", "P-2", "statement")
	again.Digest = "sha256:D-1"
	addAll(t, r, document("D-1", "P-1", "statement"), document("D-2", "P-1", "annex"), again)
	got, err := r.Digests(t.Context())
	slices.Sort(got)
	if want := []string{"sha256:D-1", "sha256:D-2"}; err != nil || !slices.Equal(got, want) {
		t.Errorf("Digests() = %v, %v; want %v", got, err, want)
	}
}
