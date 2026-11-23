package outputtest

import (
	"errors"
	"slices"
	"testing"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

// TestDocumentRepository runs the contract of output.DocumentRepository, which
// reads what the procedure repository of the same stores records. newStores
// returns empty stores each time it is called.
func TestDocumentRepository(t *testing.T, newStores func(t *testing.T) ProcedureStores) {
	t.Run("in the order they were added", func(t *testing.T) { inOrder(t, published(t, newStores(t))) })
	t.Run("only in their procedure", func(t *testing.T) { onlyInTheirProcedure(t, published(t, newStores(t))) })
	t.Run("digests of every procedure", func(t *testing.T) { digestsOfEveryProcedure(t, published(t, newStores(t))) })
}

// startWith starts the procedure with the first document and attaches the
// rest.
func startWith(t *testing.T, s ProcedureStores, procedureID string, documents ...domain.Document) {
	t.Helper()
	add(t, s, output.Start{Procedure: procedure(procedureID, "documents"), Documents: documents[:1]})
	for _, d := range documents[1:] {
		if err := s.Procedures.Attach(t.Context(), d, attachedEvent(d)); err != nil {
			t.Fatalf("Attach(%s) = %v", d.ID, err)
		}
	}
}

func inOrder(t *testing.T, s ProcedureStores) {
	added := []domain.Document{
		document("D-3", "P-1", "statement"),
		document("D-1", "P-1", "annex"),
		document("D-2", "P-1", "statement"),
	}
	startWith(t, s, "P-1", added...)
	got, err := s.Documents.List(t.Context(), "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, added) {
		t.Errorf("List = %+v, want %+v", got, added)
	}
}

func onlyInTheirProcedure(t *testing.T, s ProcedureStores) {
	startWith(t, s, "P-1", document("D-1", "P-1", "statement"))
	startWith(t, s, "P-2", document("D-2", "P-2", "statement"))
	got, err := s.Documents.List(t.Context(), "P-2")
	if err != nil || len(got) != 1 || got[0].ID != "D-2" {
		t.Errorf("List(P-2) = %+v, %v; want only D-2", got, err)
	}
	if empty, err := s.Documents.List(t.Context(), "P-3"); err != nil || len(empty) != 0 {
		t.Errorf("List(P-3) = %+v, %v; want none", empty, err)
	}
	if d, err := s.Documents.Get(t.Context(), "P-1", "D-1"); err != nil || d != document("D-1", "P-1", "statement") {
		t.Errorf("Get(P-1, D-1) = %+v, %v", d, err)
	}
	if _, err := s.Documents.Get(t.Context(), "P-2", "D-1"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("Get(P-2, D-1) = %v, want ErrDocumentNotFound", err)
	}
}

func digestsOfEveryProcedure(t *testing.T, s ProcedureStores) {
	again := document("D-3", "P-2", "statement")
	again.Digest = "sha256:D-1"
	startWith(t, s, "P-1", document("D-1", "P-1", "statement"), document("D-2", "P-1", "annex"))
	startWith(t, s, "P-2", again)
	got, err := s.Documents.Digests(t.Context())
	slices.Sort(got)
	if want := []string{"sha256:D-1", "sha256:D-2"}; err != nil || !slices.Equal(got, want) {
		t.Errorf("Digests() = %v, %v; want %v", got, err, want)
	}
}
