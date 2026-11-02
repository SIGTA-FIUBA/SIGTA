package rest_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/src/adapters/rest"
	"backend/src/domain/procedure"
	"backend/src/ports/input"
)

type fakeCatalog struct {
	input.ProcedureTypes
	entries []input.CatalogEntry
	err     error
}

func (f fakeCatalog) Catalog(context.Context) ([]input.CatalogEntry, error) {
	return f.entries, f.err
}

func getCatalog(t *testing.T, catalog fakeCatalog) *httptest.ResponseRecorder {
	t.Helper()
	api := rest.NewServer(nil, rest.NewCatalogHandler(catalog))
	handler := rest.Mount(http.NewServeMux(), api, slog.New(slog.DiscardHandler))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedure-types", http.NoBody))
	return recorder
}

func TestListProcedureTypesNeedsNoSession(t *testing.T) {
	recorder := getCatalog(t, fakeCatalog{entries: []input.CatalogEntry{{
		ProcedureType: 7,
		Version:       2,
		Name:          "Trámite",
		Description:   "Descripción",
		Fields:        []procedure.FieldRequirement{{Field: procedure.Field{ID: "reason", Name: "Motivo"}, Required: true}},
		Documents: []procedure.DocumentRequirement{{
			Document: procedure.Document{ID: "letter", Name: "Carta", AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024},
		}},
	}}})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var got []rest.CatalogEntry
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Id != 7 || got[0].Version != 2 || !got[0].Fields[0].Required {
		t.Errorf("body = %s", recorder.Body)
	}
	if document := got[0].Documents[0]; document.Required || document.AcceptedTypes[0] != "application/pdf" || document.MaxSize != 1024 {
		t.Errorf("document = %+v", document)
	}
	// Required lists are never null, even when the configuration has none.
	if !strings.Contains(recorder.Body.String(), `"deadlines":[]`) || !strings.Contains(recorder.Body.String(), `"admissibility":[]`) {
		t.Errorf("body = %s, want empty lists", recorder.Body)
	}
}

func TestListProcedureTypesEmpty(t *testing.T) {
	if body := strings.TrimSpace(getCatalog(t, fakeCatalog{}).Body.String()); body != "[]" {
		t.Errorf("body = %s, want []", body)
	}
}

func TestListProcedureTypesFailure(t *testing.T) {
	recorder := getCatalog(t, fakeCatalog{err: errors.New("store unavailable")})
	if recorder.Code != http.StatusInternalServerError || strings.Contains(recorder.Body.String(), "store") {
		t.Errorf("status = %d, body = %s", recorder.Code, recorder.Body)
	}
}
