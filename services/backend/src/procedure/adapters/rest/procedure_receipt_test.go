package rest_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/encoding/charmap"

	"backend/src/api"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
)

func downloadReceipt(t *testing.T, procedures fakeProcedures, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	server := api.NewServer(nil, nil, nil, nil, rest.NewProcedureHandler(fakeAuth{}, fakePeople{}, procedures, nil, nil), nil, nil)
	handler := api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/procedures/P-1/receipt", nil)
	r.Header.Set("Cookie", "sigta_session="+cookie)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)
	return recorder
}

// shows says whether the PDF writes text on a line of its own or as part of
// one, as the PDF encodes it.
func shows(t *testing.T, document []byte, text string) bool {
	t.Helper()
	encoded, err := charmap.Windows1252.NewEncoder().String(text)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Contains(document, []byte(strings.NewReplacer(`(`, `\(`, `)`, `\)`).Replace(encoded)))
}

func receiptOf(detail input.ProcedureDetail) input.Receipt {
	return input.Receipt{ProcedureDetail: detail, IssuedAt: time.Date(2026, 11, 20, 15, 30, 0, 0, time.UTC)}
}

func TestDownloadProcedureReceiptIsAPDFToSave(t *testing.T) {
	recorder := downloadReceipt(t, fakeProcedures{receipt: receiptOf(openedDetail(false))}, session)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	h := recorder.Header()
	if h.Get("Content-Type") != "application/pdf" || h.Get("Content-Disposition") != "attachment; filename=comprobante-P-1.pdf" || h.Get("Cache-Control") != "no-store" {
		t.Errorf("headers = %v", h)
	}
	if !bytes.HasPrefix(recorder.Body.Bytes(), []byte("%PDF-")) {
		t.Errorf("body = %q, want a PDF", recorder.Body.Bytes())
	}
}

func TestDownloadProcedureReceiptShowsTheProcedureAndItsRoute(t *testing.T) {
	d := openedDetail(false)
	d.Procedure.StartedAt = time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	d.Procedure.MovedAt = time.Date(2026, 11, 18, 2, 15, 0, 0, time.UTC)
	d.History = []domain.Event{
		{Step: "documents", By: "ana", At: d.Procedure.StartedAt, Next: "review"},
		{Step: "review", At: time.Date(2026, 11, 17, 13, 0, 0, 0, time.UTC), SendBack: true, Next: "documents", Observation: "falta el programa"},
		{Step: "documents", By: "ana", At: d.Procedure.MovedAt, Document: &domain.DocumentRecord{ID: "D-2", Requirement: "transcript", FileName: "historia (2).pdf"}},
	}

	body := downloadReceipt(t, fakeProcedures{receipt: receiptOf(d)}, session).Body.Bytes()

	for _, want := range []string{
		"Comprobante de trámite",
		"P-1",
		"Equivalencias (versión 2)",
		"Ana Prueba",
		"Pendiente del titular",
		"Adjuntar documentos · Titular",
		"16/11/2026 09:00",
		"17/11/2026 23:15",
		"20/11/2026 12:30",
		"Recorrido",
		"17/11/2026 10:00 · Revisar",
		"Dirección de Alumnos",
		"Devuelto a Adjuntar documentos",
		"Motivo: falta el programa",
		"17/11/2026 23:15 · Documento adjuntado en Adjuntar documentos",
		"Historia académica: historia (2).pdf",
		"No es una resolución",
	} {
		if !shows(t, body, want) {
			t.Errorf("the receipt does not show %q", want)
		}
	}
}

func TestDownloadProcedureReceiptOfAClosedProcedureHasNoOpenStep(t *testing.T) {
	d := openedDetail(false)
	d.Procedure.OpenStep, d.State = "", domain.StateClosed

	body := downloadReceipt(t, fakeProcedures{receipt: receiptOf(d)}, session).Body.Bytes()

	if !shows(t, body, "Cerrado") || shows(t, body, "Paso abierto") {
		t.Errorf("the receipt of a closed procedure shows an open step or no state")
	}
}

func TestDownloadProcedureReceiptAnswersNotFoundAndSignIn(t *testing.T) {
	if code := downloadReceipt(t, fakeProcedures{err: domain.ErrProcedureNotFound}, session).Code; code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", code)
	}
	if code := downloadReceipt(t, fakeProcedures{}, "other").Code; code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", code)
	}
}
