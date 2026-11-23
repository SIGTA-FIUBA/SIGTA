package rest_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/procedure/adapters/pdf/pdftest"
	"backend/src/procedure/adapters/rest"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
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

// pagesOf downloads the receipt and returns the lines it shows on each page.
func pagesOf(t *testing.T, r input.Receipt) [][]string {
	t.Helper()
	recorder := downloadReceipt(t, fakeProcedures{receipt: r}, session)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	return pdftest.Lines(t, recorder.Body.Bytes())
}

func receiptOf(detail input.ProcedureDetail) input.Receipt {
	return input.Receipt{ProcedureDetail: detail, IssuedAt: time.Date(2026, 11, 20, 15, 30, 0, 0, time.UTC)}
}

// evaluatedDetail is openedDetail with a step of another area, so that the
// procedure can go out for evaluation and come back.
func evaluatedDetail() input.ProcedureDetail {
	d := openedDetail(false)
	d.Steps = append(d.Steps, detail.Step{
		Step: definition.Step{ID: "evaluate", Name: "Evaluar", Action: definition.ActionDecide, Role: definition.RoleStaff, Area: "commission"},
		Area: &detail.Area{ID: "commission", Name: "Comisión Curricular"},
	})
	d.Procedure.StartedAt = time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	d.Procedure.MovedAt = time.Date(2026, 11, 19, 2, 30, 0, 0, time.UTC)
	return d
}

// header is what the receipt shows before the history.
var header = []string{
	"Comprobante de trámite",
	"SIGTA · Secretaría de Gestión Académica · Facultad de Ingeniería, Universidad de Buenos Aires",
	"Trámite", "P-1",
	"Tipo de trámite", "Equivalencias (versión 2)",
	"Titular", "Ana Prueba",
}

func TestDownloadProcedureReceiptIsAPDFToSave(t *testing.T) {
	recorder := downloadReceipt(t, fakeProcedures{receipt: receiptOf(openedDetail(false))}, session)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d: %s", recorder.Code, recorder.Body)
	}
	h := recorder.Header()
	if h.Get("Content-Type") != "application/pdf" || h.Get("Content-Disposition") != "attachment; filename=comprobante-P-1.pdf" ||
		h.Get("X-Content-Type-Options") != "nosniff" || h.Get("Content-Security-Policy") != "sandbox" || h.Get("Cache-Control") != "no-store" {
		t.Errorf("headers = %v", h)
	}
	if !bytes.HasPrefix(recorder.Body.Bytes(), []byte("%PDF-")) {
		t.Errorf("body = %q, want a PDF", recorder.Body.Bytes())
	}
}

// The history goes through every kind of event, as the procedure page names
// them, with the person on the steps of the applicant and the area on those
// of staff.
func TestDownloadProcedureReceiptShowsEveryLineOfTheProcedureAndItsHistory(t *testing.T) {
	d := evaluatedDetail()
	at := func(day, hour int) time.Time { return time.Date(2026, 11, day, hour, 0, 0, 0, time.UTC) }
	d.History = []domain.Event{
		{Step: "documents", By: "ana", At: at(16, 12), Next: "review"},
		{Step: "review", At: at(17, 13), Next: "evaluate", Referral: &domain.ReferralRecord{From: "office", To: "commission"}},
		{Step: "evaluate", At: at(17, 14), Observation: "Falta la firma.", CorrectionRequest: &domain.CorrectionRecord{Fields: []string{"reason"}, Documents: []string{"transcript"}}},
		{Step: "documents", By: "ana", At: at(18, 9), Document: &domain.DocumentRecord{ID: "D-2", Requirement: "transcript", FileName: `historia (2) \ firmada.pdf`}},
		{Step: "documents", By: "ana", At: at(18, 9), Value: &domain.ValueRecord{Field: "reason"}},
		{Step: "evaluate", By: "ana", At: at(18, 10), CorrectionAnswer: &domain.CorrectionRecord{Fields: []string{"reason"}, Documents: []string{"transcript"}}},
		{Step: "evaluate", At: at(18, 15), Next: "review", Returned: true},
		{Step: "review", At: at(19, 2).Add(30 * time.Minute), Next: "documents", SendBack: true, Declared: true, Observation: "Falta el programa."},
	}

	pages := pagesOf(t, receiptOf(d))

	want := append(slices.Clone(header),
		"Estado", "Pendiente del titular",
		"Paso actual", "Adjuntar documentos · Titular",
		"Iniciado el", "16/11/2026 09:00",
		"Último movimiento", "18/11/2026 23:30",
		"Emitido el", "20/11/2026 12:30",
		"Historial",
		"16/11/2026 09:00 · Adjuntar documentos", "Ana Prueba",
		"17/11/2026 10:00 · Revisar", "Dirección de Alumnos", "Derivado a Comisión Curricular",
		"17/11/2026 11:00 · Subsanación pedida en Evaluar", "Comisión Curricular", "Se pide: Motivo, Historia académica", "Motivo: Falta la firma.",
		"18/11/2026 06:00 · Documento adjuntado en Adjuntar documentos", "Ana Prueba", `Historia académica: historia (2) \ firmada.pdf`,
		"18/11/2026 06:00 · Dato cargado en Adjuntar documentos", "Ana Prueba", "Motivo",
		"18/11/2026 07:00 · Respuesta a la subsanación en Evaluar", "Ana Prueba", "Se respondió: Motivo, Historia académica",
		"18/11/2026 12:00 · Evaluar", "Comisión Curricular", "Volvió al área de la que se derivó",
		"18/11/2026 23:30 · Revisar", "Dirección de Alumnos", "Devuelto a Adjuntar documentos", "Motivo: Falta el programa.", "Hecho fuera de SIGTA, declarado por quien lo confirmó",
	)
	if len(pages) != 1 {
		t.Fatalf("%d pages, want 1", len(pages))
	}
	got := pages[0]
	if len(got) < len(want) || fmt.Sprint(got[:len(want)]) != fmt.Sprint(want) {
		t.Fatalf("lines:\n%s\nwant them to start with:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	rest := strings.Join(got[len(want):], " ")
	if !strings.HasPrefix(rest, "Este comprobante deja constancia") || !strings.Contains(rest, "No es una resolución") ||
		!strings.HasSuffix(rest, "SIGTA · Comprobante del trámite P-1 · Página 1 de 1") {
		t.Errorf("the receipt ends with %q", rest)
	}
}

func TestDownloadProcedureReceiptSaysWhereAReferredProcedureIs(t *testing.T) {
	d := evaluatedDetail()
	d.Procedure.OpenStep, d.State = "evaluate", domain.StateReferred
	d.Referral = &detail.Referral{Area: detail.Area{ID: "commission", Name: "Comisión Curricular"}, From: detail.Area{ID: "office", Name: "Dirección de Alumnos"}, DaysWithoutReturn: 1}

	got := pagesOf(t, receiptOf(d))[0]

	want := append(slices.Clone(header),
		"Estado", "Derivado",
		"Paso actual", "Evaluar · Comisión Curricular",
		"Derivación", "En Comisión Curricular, desde Dirección de Alumnos. 1 día sin retorno.",
		"Iniciado el",
	)
	if fmt.Sprint(got[:len(want)]) != fmt.Sprint(want) {
		t.Errorf("lines = %q, want them to start with %q", got[:len(want)], want)
	}
}

func TestDownloadProcedureReceiptOfAClosedProcedureSaysWhenItClosed(t *testing.T) {
	d := evaluatedDetail()
	d.Procedure.OpenStep, d.State = "", domain.StateClosed

	got := pagesOf(t, receiptOf(d))[0]

	want := append(slices.Clone(header),
		"Estado", "Cerrado",
		"Iniciado el", "16/11/2026 09:00",
		"Cerrado el", "18/11/2026 23:30",
		"Emitido el", "20/11/2026 12:30",
		"Historial",
		"Todavía no hay pasos hechos.",
	)
	if fmt.Sprint(got[:len(want)]) != fmt.Sprint(want) {
		t.Errorf("lines = %q, want them to start with %q", got[:len(want)], want)
	}
}

func TestDownloadProcedureReceiptNamesEachState(t *testing.T) {
	for state, name := range map[domain.State]string{
		domain.StateWithApplicant: "Pendiente del titular",
		domain.StateInOffice:      "En un área",
		domain.StateReferred:      "Derivado",
		domain.StateClosed:        "Cerrado",
	} {
		d := openedDetail(false)
		d.State = state
		if got := pagesOf(t, receiptOf(d))[0]; got[len(header)] != "Estado" || got[len(header)+1] != name {
			t.Errorf("%s: lines %q, want the state %q", state, got[len(header):len(header)+2], name)
		}
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
