package rest

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"strings"
	"time"

	"backend/src/api"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/pdf"
	"backend/src/procedure/domain"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/ports/input"
)

// argentina is the time of the Secretaría, which has had no daylight saving
// since 2009. A fixed zone does not depend on the time zones of the server.
var argentina = time.FixedZone("ART", -3*60*60)

var stateNames = map[domain.State]string{
	domain.StateInOffice:      "En un área",
	domain.StateReferred:      "Derivado",
	domain.StateWithApplicant: "Pendiente del titular",
	domain.StateClosed:        "Cerrado",
}

func (h *ProcedureHandler) DownloadProcedureReceipt(ctx context.Context, req api.DownloadProcedureReceiptRequestObject) (api.DownloadProcedureReceiptResponseObject, error) {
	user, err := h.auth.Authenticate(ctx, valueOf(req.Params.SigtaSession))
	if authinput.Unauthenticated(err) {
		return api.DownloadProcedureReceipt401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse{Message: messageSignIn}}, nil
	}
	if err != nil {
		return nil, err
	}
	r, err := h.procedures.Receipt(ctx, user.ID, req.ProcedureId)
	if errors.Is(err, domain.ErrProcedureNotFound) {
		return api.DownloadProcedureReceipt404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: messageNoProcedure}}, nil
	}
	if err != nil {
		return nil, err
	}
	people := []string{r.Procedure.Applicant}
	for _, e := range r.History {
		people = append(people, e.By)
	}
	names, err := h.people.Names(ctx, people)
	if err != nil {
		return nil, err
	}
	content := receiptPDF(r, names)
	return api.DownloadProcedureReceipt200ApplicationpdfResponse{
		Body:          strings.NewReader(string(content)),
		ContentLength: int64(len(content)),
		Headers: api.DownloadProcedureReceipt200ResponseHeaders{
			ContentDisposition:    mime.FormatMediaType(dispositionAttachment, map[string]string{"filename": "comprobante-" + r.Procedure.ID + ".pdf"}),
			XContentTypeOptions:   noSniff,
			ContentSecurityPolicy: sandbox,
			CacheControl:          noStore,
		},
	}, nil
}

// receiptPDF lays out the receipt with what the web shows of the procedure to
// its applicant, in the same words.
func receiptPDF(r input.Receipt, names map[string]string) []byte {
	p := r.Procedure
	doc := pdf.New("Comprobante del trámite "+p.ID, "SIGTA · Comprobante del trámite "+p.ID)
	doc.Title("Comprobante de trámite")
	doc.Note("SIGTA · Secretaría de Gestión Académica · Facultad de Ingeniería, Universidad de Buenos Aires")
	doc.Space()
	doc.Field("Trámite", p.ID)
	doc.Field("Tipo de trámite", fmt.Sprintf("%s (versión %d)", r.Name, p.Version))
	doc.Field("Titular", nameOrID(names[p.Applicant], p.Applicant))
	doc.Field("Estado", stateNames[r.State])
	if step := stepOf(r.Steps, p.OpenStep); step != nil {
		doc.Field("Paso actual", step.Step.Name+" · "+responsibleOf(step))
	}
	if ref := r.Referral; ref != nil {
		doc.Field("Derivación", referralText(*ref))
	}
	doc.Field("Iniciado el", dateTime(p.StartedAt))
	if p.Closed() {
		doc.Field("Cerrado el", dateTime(p.MovedAt))
	} else {
		doc.Field("Último movimiento", dateTime(p.MovedAt))
	}
	doc.Field("Emitido el", dateTime(r.IssuedAt))

	doc.Heading("Historial")
	if len(r.History) == 0 {
		doc.Text("Todavía no hay pasos hechos.")
	}
	for _, e := range r.History {
		historyEvent(doc, r, e, names)
	}
	doc.Space()
	doc.Note("Este comprobante deja constancia del estado y del recorrido del trámite en SIGTA cuando se emitió. No es una resolución ni reemplaza a la que se adjunte al trámite. Horarios de Argentina.")
	return doc.Bytes()
}

func historyEvent(doc *pdf.Document, r input.Receipt, e domain.Event, names map[string]string) {
	step := stepOf(r.Steps, e.Step)
	name := e.Step
	if step != nil {
		name = step.Step.Name
	}
	title := name
	switch {
	case e.Document != nil:
		title = "Documento adjuntado en " + name
	case e.Value != nil:
		title = "Dato cargado en " + name
	case e.CorrectionAnswer != nil:
		title = "Respuesta a la subsanación en " + name
	case e.CorrectionRequest != nil:
		title = "Subsanación pedida en " + name
	}
	doc.Space()
	doc.Bold(dateTime(e.At) + " · " + title)
	by := "Personal"
	switch {
	case e.By != "":
		by = nameOrID(names[e.By], e.By)
	case step != nil:
		by = responsibleOf(step)
	}
	doc.Indented(by)
	if d := e.Document; d != nil {
		doc.Indented(fmt.Sprintf("%s: %s", documentName(r.Steps, d.Requirement), d.FileName))
	}
	if e.CorrectionRequest != nil {
		doc.Indented("Se pide: " + requestedNames(r.Steps, *e.CorrectionRequest))
	}
	if e.Value != nil {
		doc.Indented(requestedNames(r.Steps, domain.CorrectionRecord{Fields: []string{e.Value.Field}}))
	}
	if e.CorrectionAnswer != nil {
		doc.Indented("Se respondió: " + requestedNames(r.Steps, *e.CorrectionAnswer))
	}
	if e.SendBack {
		next := e.Next
		if s := stepOf(r.Steps, e.Next); s != nil {
			next = s.Step.Name
		}
		doc.Indented("Devuelto a " + next)
	}
	if e.Observation != "" {
		if e.SendBack || e.CorrectionRequest != nil {
			doc.Indented("Motivo: " + e.Observation)
		} else {
			doc.Indented(e.Observation)
		}
	}
	if e.Referral != nil {
		doc.Indented("Derivado a " + areaName(r.Steps, e.Referral.To))
	}
	if e.Returned {
		doc.Indented("Volvió al área de la que se derivó")
	}
	if e.Declared {
		doc.Indented("Hecho fuera de SIGTA, declarado por quien lo confirmó")
	}
}

// referralText says where a referred procedure is and for how long, as the
// procedure page does.
func referralText(r detail.Referral) string {
	days := fmt.Sprintf("%d días", r.DaysWithoutReturn)
	if r.DaysWithoutReturn == 1 {
		days = "1 día"
	}
	return fmt.Sprintf("En %s, desde %s. %s sin retorno.", r.Area.Name, r.From.Name, days)
}

func dateTime(t time.Time) string {
	return t.In(argentina).Format("02/01/2006 15:04")
}

func stepOf(steps []detail.Step, id string) *detail.Step {
	for i := range steps {
		if steps[i].Step.ID == id {
			return &steps[i]
		}
	}
	return nil
}

// responsibleOf says who has a step: its area or, if it has none, the
// applicant.
func responsibleOf(s *detail.Step) string {
	if s.Area != nil {
		return s.Area.Name
	}
	return "Titular"
}

func areaName(steps []detail.Step, id string) string {
	for _, s := range steps {
		if s.Area != nil && s.Area.ID == id {
			return s.Area.Name
		}
	}
	return id
}

func documentName(steps []detail.Step, id string) string {
	for _, s := range steps {
		for _, d := range s.Documents {
			if d.Document.ID == id {
				return d.Document.Name
			}
		}
	}
	return id
}

func fieldName(steps []detail.Step, id string) string {
	for _, s := range steps {
		for _, f := range s.Fields {
			if f.Field.ID == id {
				return f.Field.Name
			}
		}
	}
	return id
}

func requestedNames(steps []detail.Step, asked domain.CorrectionRecord) string {
	var named []string
	for _, f := range asked.Fields {
		named = append(named, fieldName(steps, f))
	}
	for _, d := range asked.Documents {
		named = append(named, documentName(steps, d))
	}
	return strings.Join(named, ", ")
}
