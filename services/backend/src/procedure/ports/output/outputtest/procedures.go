package outputtest

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/output"
)

// ProcedureStores are the repositories of procedures, which follow a
// published version, and of the documents they record, which read the same
// store.
type ProcedureStores struct {
	Versions   output.ProcedureTypeVersionRepository
	Procedures output.ProcedureRepository
	Documents  output.DocumentRepository
}

// TestProcedureRepository runs the contract of output.ProcedureRepository.
// newStores returns empty stores each time it is called.
func TestProcedureRepository(t *testing.T, newStores func(t *testing.T) ProcedureStores) {
	t.Run("start", func(t *testing.T) { startRoundTrip(t, published(t, newStores(t))) })
	t.Run("start twice", func(t *testing.T) { startTwice(t, published(t, newStores(t))) })
	t.Run("start whole or nothing", func(t *testing.T) { startWholeOrNothing(t, published(t, newStores(t))) })
	t.Run("ordered by identifier", func(t *testing.T) { orderedByIdentifier(t, published(t, newStores(t))) })
	t.Run("advance", func(t *testing.T) { advance(t, published(t, newStores(t))) })
	t.Run("advance from a step that is not open", func(t *testing.T) { advanceFromAStepNotOpen(t, published(t, newStores(t))) })
	t.Run("unknown procedure", func(t *testing.T) { unknownProcedure(t, published(t, newStores(t))) })
	t.Run("attach", func(t *testing.T) { attach(t, published(t, newStores(t))) })
	t.Run("attach whole or nothing", func(t *testing.T) { attachWholeOrNothing(t, published(t, newStores(t))) })
	t.Run("attach to a step that is not open", func(t *testing.T) { attachToAStepNotOpen(t, published(t, newStores(t))) })
	t.Run("request correction", func(t *testing.T) { requestCorrection(t, published(t, newStores(t))) })
	t.Run("request correction on a step that is not open", func(t *testing.T) { requestCorrectionOnAStepNotOpen(t, published(t, newStores(t))) })
	t.Run("advance while the applicant has a correction to make", func(t *testing.T) { advanceAwaitingApplicant(t, published(t, newStores(t))) })
	t.Run("no shared memory", func(t *testing.T) { procedureCopies(t, published(t, newStores(t))) })
	t.Run("no shared memory in correction requests", func(t *testing.T) { correctionCopies(t, published(t, newStores(t))) })
	t.Run("concurrent confirmations", func(t *testing.T) { concurrentConfirmations(t, published(t, newStores(t))) })
	t.Run("confirmation and correction request at once", func(t *testing.T) { concurrentConfirmationAndCorrection(t, published(t, newStores(t))) })
	t.Run("concurrent attachments", func(t *testing.T) { concurrentAttachments(t, published(t, newStores(t))) })
	t.Run("enter a value in answer", func(t *testing.T) { enterValue(t, published(t, newStores(t))) })
	t.Run("attach in answer", func(t *testing.T) { attachInAnswer(t, published(t, newStores(t))) })
	t.Run("answer a correction", func(t *testing.T) { answerCorrection(t, published(t, newStores(t))) })
	t.Run("answer read before another request", func(t *testing.T) { answerBeforeAnotherRequest(t, published(t, newStores(t))) })
}

// startedAt has microseconds, the most PostgreSQL keeps.
var startedAt = time.Date(2026, 11, 14, 12, 0, 0, 123456000, time.UTC)

// published adds the version that the procedures of these tests follow.
func published(t *testing.T, s ProcedureStores) ProcedureStores {
	t.Helper()
	if err := s.Versions.Add(t.Context(), Version(1, 1, "Trámite")); err != nil {
		t.Fatalf("Add() of the version = %v", err)
	}
	return s
}

func procedure(id, openStep string) domain.Procedure {
	return domain.Procedure{
		ID:            id,
		ProcedureType: 1,
		Version:       1,
		VersionDigest: Version(1, 1, "Trámite").Digest,
		Applicant:     "titular@fi.uba.ar",
		OpenStep:      openStep,
		StartedAt:     startedAt,
		MovedAt:       startedAt,
	}
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
		UploadedAt:  startedAt,
	}
}

func attachedEvent(d domain.Document) domain.Event {
	return domain.Event{
		ProcedureID: d.Procedure,
		Step:        d.Step,
		Action:      definition.ActionUploadDocument,
		By:          d.UploadedBy,
		At:          d.UploadedAt,
		Document:    &domain.DocumentRecord{ID: d.ID, Requirement: d.Requirement, FileName: d.FileName},
	}
}

// correction asks for fields and documents on step, as the engine does: the
// reason is the observation of the event.
func correction(procedureID, step string, fields, documents []string) (domain.CorrectionRequest, domain.Event) {
	r := domain.CorrectionRequest{Step: step, By: "staff@fi.uba.ar", At: startedAt.Add(30 * time.Minute), Reason: "Falta la firma", Fields: fields, Documents: documents}
	return r, domain.Event{
		ProcedureID:       procedureID,
		Step:              step,
		By:                r.By,
		At:                r.At,
		Observation:       r.Reason,
		CorrectionRequest: &domain.CorrectionRecord{Fields: slices.Clone(fields), Documents: slices.Clone(documents)},
	}
}

func requestCorrectionOf(t *testing.T, s ProcedureStores, r domain.CorrectionRequest, e domain.Event) {
	t.Helper()
	if err := s.Procedures.RequestCorrection(t.Context(), r, e); err != nil {
		t.Fatalf("RequestCorrection(%s, %s) = %v", e.ProcedureID, r.Step, err)
	}
}

func confirmation(procedureID, step, next string) domain.Event {
	return domain.Event{ProcedureID: procedureID, Step: step, Action: definition.ActionConfirmExternalFact, By: "staff@fi.uba.ar", At: startedAt.Add(time.Hour), Next: next}
}

// fullStart fills every field the start records, each with a value that
// tells it apart.
func fullStart(id string) output.Start {
	return output.Start{
		Procedure: procedure(id, "documents"),
		Events: []domain.Event{
			{ProcedureID: id, Step: "request", Action: definition.ActionEnterData, By: "titular@fi.uba.ar", At: startedAt, Observation: "Pido la equivalencia", Next: "declare"},
			{ProcedureID: id, Step: "declare", Action: definition.ActionDecide, By: "titular@fi.uba.ar", At: startedAt, Outcome: "accept", Declared: true, Next: "documents"},
		},
		Values: []domain.FieldValue{
			{Procedure: id, Field: "reason", Step: "request", Text: "Cambio de plan, año 2026", EnteredBy: "titular@fi.uba.ar", EnteredAt: startedAt},
			{Procedure: id, Field: "subjects", Step: "request", Items: []string{"Análisis II", "Física I"}, EnteredBy: "titular@fi.uba.ar", EnteredAt: startedAt},
		},
		Documents: []domain.Document{document(id+"-D-2", id, "statement"), document(id+"-D-1", id, "annex")},
	}
}

func add(t *testing.T, s ProcedureStores, start output.Start) {
	t.Helper()
	if err := s.Procedures.Add(t.Context(), start); err != nil {
		t.Fatalf("Add(%s) = %v", start.Procedure.ID, err)
	}
}

// wantRecorded checks everything the store returns of the procedure against
// what was recorded.
func wantRecorded(t *testing.T, s ProcedureStores, p domain.Procedure, events []domain.Event, values []domain.FieldValue, documents []domain.Document) {
	t.Helper()
	ctx := t.Context()
	got, err := s.Procedures.Get(ctx, p.ID)
	if err != nil || !reflect.DeepEqual(got, p) {
		t.Errorf("Get(%s) = %+v, %v; want %+v", p.ID, got, err, p)
	}
	if err == nil && (got.StartedAt.Location() != time.UTC || got.MovedAt.Location() != time.UTC) {
		t.Errorf("Get(%s) returned times in %s and %s, want UTC", p.ID, got.StartedAt.Location(), got.MovedAt.Location())
	}
	gotEvents, err := s.Procedures.Events(ctx, p.ID)
	if err != nil || len(gotEvents) != len(events) || len(events) > 0 && !reflect.DeepEqual(gotEvents, events) {
		t.Errorf("Events(%s) = %+v, %v; want %+v", p.ID, gotEvents, err, events)
	}
	gotValues, err := s.Procedures.Values(ctx, p.ID)
	if err != nil || len(gotValues) != len(values) || len(values) > 0 && !reflect.DeepEqual(gotValues, values) {
		t.Errorf("Values(%s) = %+v, %v; want %+v", p.ID, gotValues, err, values)
	}
	gotDocuments, err := s.Documents.List(ctx, p.ID)
	if err != nil || len(gotDocuments) != len(documents) || len(documents) > 0 && !slices.Equal(gotDocuments, documents) {
		t.Errorf("List(%s) = %+v, %v; want %+v", p.ID, gotDocuments, err, documents)
	}
}

func startRoundTrip(t *testing.T, s ProcedureStores) {
	start := fullStart("P-1")
	add(t, s, start)
	wantRecorded(t, s, start.Procedure, start.Events, start.Values, start.Documents)
	for _, d := range start.Documents {
		if got, err := s.Documents.Get(t.Context(), "P-1", d.ID); err != nil || got != d {
			t.Errorf("Get(P-1, %s) = %+v, %v; want %+v", d.ID, got, err, d)
		}
	}

	empty := output.Start{Procedure: procedure("P-2", "request")}
	add(t, s, empty)
	wantRecorded(t, s, empty.Procedure, nil, nil, nil)
}

func startTwice(t *testing.T, s ProcedureStores) {
	start := fullStart("P-1")
	add(t, s, start)
	again := fullStart("P-1")
	again.Procedure.Applicant = "otro@fi.uba.ar"
	again.Documents = []domain.Document{document("D-9", "P-1", "statement")}
	if err := s.Procedures.Add(t.Context(), again); !errors.Is(err, output.ErrProcedureExists) {
		t.Fatalf("Add() of the same procedure = %v, want ErrProcedureExists", err)
	}
	wantRecorded(t, s, start.Procedure, start.Events, start.Values, start.Documents)
}

// startWholeOrNothing makes the last document of a start fail, after the
// procedure, its events, its values and its other documents. A start that
// failed can be done again.
func startWholeOrNothing(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	add(t, s, fullStart("P-1"))

	for _, tc := range []struct {
		name      string
		documents []domain.Document
	}{
		{"a document of another procedure", append(fullStart("P-2").Documents, document("P-1-D-1", "P-2", "annex"))},
		{"a document twice", append(fullStart("P-2").Documents, document("P-2-D-2", "P-2", "annex"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := fullStart("P-2")
			start.Documents = tc.documents
			if err := s.Procedures.Add(ctx, start); !errors.Is(err, output.ErrDocumentExists) {
				t.Fatalf("Add() = %v, want ErrDocumentExists", err)
			}
			if _, err := s.Procedures.Get(ctx, "P-2"); !errors.Is(err, domain.ErrProcedureNotFound) {
				t.Errorf("Get(P-2) = %v, want ErrProcedureNotFound", err)
			}
			if values, err := s.Procedures.Values(ctx, "P-2"); err != nil || len(values) != 0 {
				t.Errorf("Values(P-2) = %+v, %v; want none", values, err)
			}
			if _, err := s.Procedures.Events(ctx, "P-2"); !errors.Is(err, domain.ErrProcedureNotFound) {
				t.Errorf("Events(P-2) = %v, want ErrProcedureNotFound", err)
			}
			if documents, err := s.Documents.List(ctx, "P-2"); err != nil || len(documents) != 0 {
				t.Errorf("List(P-2) = %+v, %v; want none", documents, err)
			}
			digests, err := s.Documents.Digests(ctx)
			slices.Sort(digests)
			if want := []string{"sha256:P-1-D-1", "sha256:P-1-D-2"}; err != nil || !slices.Equal(digests, want) {
				t.Errorf("Digests() = %v, %v; want %v", digests, err, want)
			}
		})
	}
	add(t, s, fullStart("P-2"))
}

func orderedByIdentifier(t *testing.T, s ProcedureStores) {
	for _, id := range []string{"P-2", "P-10", "p-1", "P-1"} {
		add(t, s, output.Start{Procedure: procedure(id, "request")})
	}
	got, err := s.Procedures.List(t.Context())
	var ids []string
	for _, p := range got {
		ids = append(ids, p.ID)
	}
	if want := []string{"P-1", "P-10", "P-2", "p-1"}; err != nil || !slices.Equal(ids, want) {
		t.Errorf("List() = %v, %v; want %v", ids, err, want)
	}
}

func advance(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	add(t, s, fullStart("P-2"))

	referred := procedure("P-1", "evaluate")
	referred.MovedAt = startedAt.Add(time.Hour)
	referred.Referral = &domain.Referral{From: "office", To: "committee", Since: startedAt.Add(time.Hour)}
	sent := confirmation("P-1", "documents", "evaluate")
	sent.Action = definition.ActionDecide
	sent.Outcome = "refer"
	sent.Referral = &domain.ReferralRecord{From: "office", To: "committee", ExternalFileNumber: "EX-2026-00000123"}
	if err := s.Procedures.Advance(ctx, referred, sent); err != nil {
		t.Fatalf("Advance(documents) = %v", err)
	}

	back := procedure("P-1", "documents")
	back.MovedAt = startedAt.Add(2 * time.Hour)
	returned := confirmation("P-1", "evaluate", "documents")
	returned.Returned = true
	returned.SendBack = true
	returned.Observation = "Falta la firma"
	if err := s.Procedures.Advance(ctx, back, returned); err != nil {
		t.Fatalf("Advance(evaluate) = %v", err)
	}

	closed := procedure("P-1", "")
	closed.MovedAt = startedAt.Add(3 * time.Hour)
	if err := s.Procedures.Advance(ctx, closed, confirmation("P-1", "documents", "")); err != nil {
		t.Fatalf("Advance() that closes = %v", err)
	}

	events := append(slices.Clone(start.Events), sent, returned, confirmation("P-1", "documents", ""))
	wantRecorded(t, s, closed, events, start.Values, start.Documents)
	other := fullStart("P-2")
	wantRecorded(t, s, other.Procedure, other.Events, other.Values, other.Documents)
}

func advanceFromAStepNotOpen(t *testing.T, s ProcedureStores) {
	start := fullStart("P-1")
	add(t, s, start)
	closed := procedure("P-2", "")
	add(t, s, output.Start{Procedure: closed})

	for _, tc := range []struct {
		procedure domain.Procedure
		step      string
	}{
		{procedure("P-1", ""), "approve"},
		{procedure("P-2", "approve"), "documents"},
	} {
		if err := s.Procedures.Advance(t.Context(), tc.procedure, confirmation(tc.procedure.ID, tc.step, "")); !errors.Is(err, engine.ErrStepNotOpen) {
			t.Errorf("Advance(%s, %s) = %v, want ErrStepNotOpen", tc.procedure.ID, tc.step, err)
		}
	}
	wantRecorded(t, s, start.Procedure, start.Events, start.Values, start.Documents)
	wantRecorded(t, s, closed, nil, nil, nil)
}

func requestCorrection(t *testing.T, s ProcedureStores) {
	start := fullStart("P-1")
	add(t, s, start)
	add(t, s, fullStart("P-2"))
	first, firstEvent := correction("P-1", "documents", []string{"reason", "subjects"}, nil)
	second, secondEvent := correction("P-1", "documents", nil, []string{"statement"})
	second.Reason, secondEvent.Observation = "Falta el anexo", "Falta el anexo"
	requestCorrectionOf(t, s, first, firstEvent)
	requestCorrectionOf(t, s, second, secondEvent)

	want := start.Procedure
	want.CorrectionRequests = []domain.CorrectionRequest{first, second}
	wantRecorded(t, s, want, append(slices.Clone(start.Events), firstEvent, secondEvent), start.Values, start.Documents)
	other := fullStart("P-2")
	wantRecorded(t, s, other.Procedure, other.Events, other.Values, other.Documents)
	if all, err := s.Procedures.List(t.Context()); err != nil || !reflect.DeepEqual(all, []domain.Procedure{want, other.Procedure}) {
		t.Errorf("List() = %+v, %v; want %+v", all, err, []domain.Procedure{want, other.Procedure})
	}
}

func requestCorrectionOnAStepNotOpen(t *testing.T, s ProcedureStores) {
	start := fullStart("P-1")
	add(t, s, start)
	closed := procedure("P-2", "")
	add(t, s, output.Start{Procedure: closed})

	for _, tc := range []struct{ procedureID, step string }{{"P-1", "approve"}, {"P-2", "documents"}} {
		r, e := correction(tc.procedureID, tc.step, []string{"reason"}, nil)
		if err := s.Procedures.RequestCorrection(t.Context(), r, e); !errors.Is(err, engine.ErrStepNotOpen) {
			t.Errorf("RequestCorrection(%s, %s) = %v, want ErrStepNotOpen", tc.procedureID, tc.step, err)
		}
	}
	wantRecorded(t, s, start.Procedure, start.Events, start.Values, start.Documents)
	wantRecorded(t, s, closed, nil, nil, nil)
}

// advanceAwaitingApplicant confirms a step read before a correction request
// was stored: the procedure does not move.
func advanceAwaitingApplicant(t *testing.T, s ProcedureStores) {
	start := fullStart("P-1")
	add(t, s, start)
	r, e := correction("P-1", "documents", []string{"reason"}, nil)
	requestCorrectionOf(t, s, r, e)

	if err := s.Procedures.Advance(t.Context(), procedure("P-1", "approve"), confirmation("P-1", "documents", "approve")); !errors.Is(err, engine.ErrAwaitingApplicant) {
		t.Errorf("Advance() = %v, want ErrAwaitingApplicant", err)
	}
	want := start.Procedure
	want.CorrectionRequests = []domain.CorrectionRequest{r}
	wantRecorded(t, s, want, append(slices.Clone(start.Events), e), start.Values, start.Documents)
}

func unknownProcedure(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	if _, err := s.Procedures.Get(ctx, "P-1"); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Get() = %v, want ErrProcedureNotFound", err)
	}
	if err := s.Procedures.Advance(ctx, procedure("P-1", ""), confirmation("P-1", "documents", "")); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Advance() = %v, want ErrProcedureNotFound", err)
	}
	if _, err := s.Procedures.Events(ctx, "P-1"); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Events() = %v, want ErrProcedureNotFound", err)
	}
	if values, err := s.Procedures.Values(ctx, "P-1"); err != nil || len(values) != 0 {
		t.Errorf("Values() = %+v, %v; want none", values, err)
	}
	if all, err := s.Procedures.List(ctx); err != nil || len(all) != 0 {
		t.Errorf("List() = %+v, %v; want none", all, err)
	}
}

func attach(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)

	d := document("D-3", "P-1", "statement")
	if err := s.Procedures.Attach(ctx, d, attachedEvent(d)); err != nil {
		t.Fatalf("Attach() = %v", err)
	}
	moved := procedure("P-1", "approve")
	if err := s.Procedures.Advance(ctx, moved, confirmation("P-1", "documents", "approve")); err != nil {
		t.Fatalf("Advance() after Attach() = %v", err)
	}

	events := append(slices.Clone(start.Events), attachedEvent(d), confirmation("P-1", "documents", "approve"))
	wantRecorded(t, s, moved, events, start.Values, append(slices.Clone(start.Documents), d))
	if got, err := s.Documents.Get(ctx, "P-1", "D-3"); err != nil || got != d {
		t.Errorf("Get(P-1, D-3) = %+v, %v; want %+v", got, err, d)
	}
}

// attachToAStepNotOpen attaches documents of a step that a confirmation
// already closed, as when both requests read the procedure before either was
// saved.
func attachToAStepNotOpen(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	closed := procedure("P-2", "")
	add(t, s, output.Start{Procedure: closed})

	another := document("D-1", "P-1", "statement")
	another.Step = "approve"
	for _, d := range []domain.Document{another, document("D-2", "P-2", "statement")} {
		if err := s.Procedures.Attach(ctx, d, attachedEvent(d)); !errors.Is(err, engine.ErrStepNotOpen) {
			t.Errorf("Attach(%s) of step %s = %v, want ErrStepNotOpen", d.Procedure, d.Step, err)
		}
	}
	wantRecorded(t, s, start.Procedure, start.Events, start.Values, start.Documents)
	wantRecorded(t, s, closed, nil, nil, nil)
}

func attachWholeOrNothing(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)

	taken := document("P-1-D-1", "P-1", "statement")
	if err := s.Procedures.Attach(ctx, taken, attachedEvent(taken)); !errors.Is(err, output.ErrDocumentExists) {
		t.Errorf("Attach() of a taken document = %v, want ErrDocumentExists", err)
	}
	orphan := document("D-9", "P-9", "statement")
	if err := s.Procedures.Attach(ctx, orphan, attachedEvent(orphan)); !errors.Is(err, domain.ErrProcedureNotFound) {
		t.Errorf("Attach() to an unknown procedure = %v, want ErrProcedureNotFound", err)
	}
	wantRecorded(t, s, start.Procedure, start.Events, start.Values, start.Documents)
	if _, err := s.Documents.Get(ctx, "P-9", "D-9"); !errors.Is(err, domain.ErrDocumentNotFound) {
		t.Errorf("Get(P-9, D-9) = %v, want ErrDocumentNotFound", err)
	}
	digests, err := s.Documents.Digests(ctx)
	if err != nil || slices.Contains(digests, "sha256:D-9") {
		t.Errorf("Digests() = %v, %v; want no digest of D-9", digests, err)
	}
}

// procedureCopies changes what was added and what was returned, through every
// slice and pointer, and expects the stored procedure unchanged.
func procedureCopies(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	referred := procedure("P-1", "evaluate")
	referred.Referral = &domain.Referral{From: "office", To: "committee", Since: startedAt}
	sent := confirmation("P-1", "documents", "evaluate")
	sent.Referral = &domain.ReferralRecord{From: "office", To: "committee"}
	if err := s.Procedures.Advance(ctx, referred, sent); err != nil {
		t.Fatal(err)
	}
	d := document("D-3", "P-1", "statement")
	d.Step = "evaluate"
	attached := attachedEvent(d)
	if err := s.Procedures.Attach(ctx, d, attached); err != nil {
		t.Fatal(err)
	}

	start.Values[1].Items[0] = "changed"
	referred.Referral.To = "changed"
	sent.Referral.To = "changed"
	attached.Document.FileName = "changed"
	got, _ := s.Procedures.Get(ctx, "P-1")
	got.Referral.From = "changed"
	events, _ := s.Procedures.Events(ctx, "P-1")
	events[2].Referral.From = "changed"
	events[3].Document.ID = "changed"
	values, _ := s.Procedures.Values(ctx, "P-1")
	values[1].Items[1] = "changed"

	want := fullStart("P-1")
	stored, _ := s.Procedures.Get(ctx, "P-1")
	if stored.Referral == nil || *stored.Referral != (domain.Referral{From: "office", To: "committee", Since: startedAt}) {
		t.Errorf("stored referral = %+v, want it unchanged", stored.Referral)
	}
	storedEvents, _ := s.Procedures.Events(ctx, "P-1")
	if len(storedEvents) != 4 || *storedEvents[2].Referral != (domain.ReferralRecord{From: "office", To: "committee"}) ||
		*storedEvents[3].Document != (domain.DocumentRecord{ID: "D-3", Requirement: "statement", FileName: "D-3.pdf"}) {
		t.Errorf("stored events = %+v, want them unchanged", storedEvents)
	}
	if storedValues, _ := s.Procedures.Values(ctx, "P-1"); !reflect.DeepEqual(storedValues, want.Values) {
		t.Errorf("stored values = %+v, want %+v", storedValues, want.Values)
	}
}

func correctionCopies(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	add(t, s, output.Start{Procedure: procedure("P-1", "documents")})
	r, e := correction("P-1", "documents", []string{"reason"}, []string{"statement"})
	requestCorrectionOf(t, s, r, e)

	r.Fields[0] = "changed"
	e.CorrectionRequest.Documents[0] = "changed"
	got, _ := s.Procedures.Get(ctx, "P-1")
	got.CorrectionRequests[0].Documents[0] = "changed"
	events, _ := s.Procedures.Events(ctx, "P-1")
	events[0].CorrectionRequest.Fields[0] = "changed"

	stored, _ := s.Procedures.Get(ctx, "P-1")
	if len(stored.CorrectionRequests) != 1 || !slices.Equal(stored.CorrectionRequests[0].Fields, []string{"reason"}) ||
		!slices.Equal(stored.CorrectionRequests[0].Documents, []string{"statement"}) {
		t.Errorf("stored requests = %+v, want them unchanged", stored.CorrectionRequests)
	}
	storedEvents, _ := s.Procedures.Events(ctx, "P-1")
	if len(storedEvents) != 1 || !reflect.DeepEqual(*storedEvents[0].CorrectionRequest, domain.CorrectionRecord{Fields: []string{"reason"}, Documents: []string{"statement"}}) {
		t.Errorf("stored events = %+v, want them unchanged", storedEvents)
	}
}

// concurrentConfirmationAndCorrection races a confirmation of the open step
// with a correction request on it: exactly one of them is recorded.
func concurrentConfirmationAndCorrection(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	for i := range 10 {
		id := fmt.Sprintf("P-%d", i)
		add(t, s, output.Start{Procedure: procedure(id, "documents")})
		var (
			wg                  sync.WaitGroup
			start               = make(chan struct{})
			advanced, requested error
		)
		wg.Go(func() {
			<-start
			advanced = s.Procedures.Advance(ctx, procedure(id, "approve"), confirmation(id, "documents", "approve"))
		})
		wg.Go(func() {
			<-start
			r, e := correction(id, "documents", []string{"reason"}, nil)
			requested = s.Procedures.RequestCorrection(ctx, r, e)
		})
		close(start)
		wg.Wait()

		got, _ := s.Procedures.Get(ctx, id)
		events, _ := s.Procedures.Events(ctx, id)
		switch {
		case advanced == nil && errors.Is(requested, engine.ErrStepNotOpen):
			if got.OpenStep != "approve" || len(got.CorrectionRequests) != 0 || len(events) != 1 {
				t.Errorf("%s advanced and then = %+v with %d events, want it on approve with no request", id, got, len(events))
			}
		case requested == nil && errors.Is(advanced, engine.ErrAwaitingApplicant):
			if got.OpenStep != "documents" || len(got.CorrectionRequests) != 1 || len(events) != 1 {
				t.Errorf("%s got a request and then = %+v with %d events, want it on documents with the request", id, got, len(events))
			}
		default:
			t.Errorf("%s: Advance() = %v and RequestCorrection() = %v, want one of them to fail", id, advanced, requested)
		}
	}
}

// concurrentConfirmations races confirmations of the same step: exactly one
// advances the procedure and the rest get ErrStepNotOpen.
func concurrentConfirmations(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	add(t, s, output.Start{Procedure: procedure("P-1", "documents")})

	const confirmations = 10
	var (
		wg    sync.WaitGroup
		start = make(chan struct{})
		errs  = make([]error, confirmations)
	)
	for i := range confirmations {
		wg.Go(func() {
			<-start
			errs[i] = s.Procedures.Advance(ctx, procedure("P-1", "approve"), confirmation("P-1", "documents", "approve"))
		})
	}
	close(start)
	wg.Wait()

	advanced := 0
	for _, err := range errs {
		switch {
		case err == nil:
			advanced++
		case !errors.Is(err, engine.ErrStepNotOpen):
			t.Errorf("Advance() = %v, want nil or ErrStepNotOpen", err)
		}
	}
	if advanced != 1 {
		t.Errorf("%d confirmations advanced the procedure, want 1", advanced)
	}
	if events, err := s.Procedures.Events(ctx, "P-1"); err != nil || len(events) != 1 {
		t.Errorf("Events() = %d events, %v; want 1", len(events), err)
	}
}

// concurrentAttachments attaches documents to the same procedure at once:
// every one is recorded with its event.
func concurrentAttachments(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	add(t, s, output.Start{Procedure: procedure("P-1", "documents")})

	const attachments = 10
	var (
		wg    sync.WaitGroup
		start = make(chan struct{})
		errs  = make([]error, attachments)
	)
	for i := range attachments {
		wg.Go(func() {
			<-start
			d := document(fmt.Sprintf("D-%d", i), "P-1", "statement")
			errs[i] = s.Procedures.Attach(ctx, d, attachedEvent(d))
		})
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("Attach(D-%d) = %v", i, err)
		}
	}
	documents, err := s.Documents.List(ctx, "P-1")
	if err != nil || len(documents) != attachments {
		t.Fatalf("List() = %d documents, %v; want %d", len(documents), err, attachments)
	}
	events, err := s.Procedures.Events(ctx, "P-1")
	if err != nil || len(events) != attachments {
		t.Fatalf("Events() = %d events, %v; want %d", len(events), err, attachments)
	}
	for i := range documents {
		if events[i].Document == nil || events[i].Document.ID != documents[i].ID {
			t.Errorf("event %d names %+v, want document %s, in the order the documents were added", i, events[i].Document, documents[i].ID)
		}
	}
}

// answered is the event of an answer to everything asked by the requests.
func answered(procedureID, step string, fields, documents []string) domain.Event {
	return domain.Event{ProcedureID: procedureID, Step: step, By: "titular@fi.uba.ar", At: startedAt.Add(2 * time.Hour),
		CorrectionAnswer: &domain.CorrectionRecord{Fields: slices.Clone(fields), Documents: slices.Clone(documents)}}
}

// enterValue records a value the applicant loads again, only while a
// correction is pending, with its event.
func enterValue(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	v := domain.FieldValue{Procedure: "P-1", Field: "reason", Step: "request", Text: "Con la firma", EnteredBy: "titular@fi.uba.ar", EnteredAt: startedAt.Add(time.Hour)}
	e := domain.Event{ProcedureID: "P-1", Step: "request", Action: definition.ActionEnterData, By: v.EnteredBy, At: v.EnteredAt, Value: &domain.ValueRecord{Field: "reason"}}
	if err := s.Procedures.EnterValue(ctx, v, e); !errors.Is(err, engine.ErrNoCorrectionPending) {
		t.Errorf("EnterValue() with nothing pending = %v, want ErrNoCorrectionPending", err)
	}
	r, requested := correction("P-1", "documents", []string{"reason"}, nil)
	requestCorrectionOf(t, s, r, requested)

	if err := s.Procedures.EnterValue(ctx, v, e); err != nil {
		t.Fatalf("EnterValue() = %v", err)
	}
	want := start.Procedure
	want.CorrectionRequests = []domain.CorrectionRequest{r}
	wantRecorded(t, s, want, append(slices.Clone(start.Events), requested, e), append(slices.Clone(start.Values), v), start.Documents)
}

// attachInAnswer attaches a document of a step of the applicant, which is not
// the open one, while a correction is pending, and refuses it once answered.
func attachInAnswer(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	r, requested := correction("P-1", "documents", nil, []string{"statement"})
	requestCorrectionOf(t, s, r, requested)

	d := document("D-3", "P-1", "statement")
	d.Step = "request"
	if err := s.Procedures.Attach(ctx, d, attachedEvent(d)); err != nil {
		t.Fatalf("Attach() in answer = %v", err)
	}
	staff := document("D-4", "P-1", "statement")
	staff.Step, staff.UploadedBy = "request", "staff@fi.uba.ar"
	if err := s.Procedures.Attach(ctx, staff, attachedEvent(staff)); !errors.Is(err, engine.ErrStepNotOpen) {
		t.Errorf("Attach() by staff to a step of the applicant = %v, want ErrStepNotOpen", err)
	}
	answer := answered("P-1", "documents", nil, []string{"statement"})
	if err := s.Procedures.AnswerCorrection(ctx, procedure("P-1", "documents"), 1, answer); err != nil {
		t.Fatalf("AnswerCorrection() = %v", err)
	}
	late := document("D-5", "P-1", "statement")
	late.Step = "request"
	if err := s.Procedures.Attach(ctx, late, attachedEvent(late)); !errors.Is(err, engine.ErrStepNotOpen) {
		t.Errorf("Attach() after the answer = %v, want ErrStepNotOpen", err)
	}
	events := append(slices.Clone(start.Events), requested, attachedEvent(d), answer)
	wantRecorded(t, s, start.Procedure, events, start.Values, append(slices.Clone(start.Documents), d))
}

// answerCorrection answers every pending request at once; a request after
// the answer is pending again, and a second answer finds nothing pending.
func answerCorrection(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	first, firstEvent := correction("P-1", "documents", []string{"reason"}, nil)
	second, secondEvent := correction("P-1", "documents", nil, []string{"statement"})
	requestCorrectionOf(t, s, first, firstEvent)
	requestCorrectionOf(t, s, second, secondEvent)

	answer := answered("P-1", "documents", []string{"reason"}, []string{"statement"})
	if err := s.Procedures.AnswerCorrection(ctx, procedure("P-1", "documents"), 2, answer); err != nil {
		t.Fatalf("AnswerCorrection() = %v", err)
	}
	if err := s.Procedures.AnswerCorrection(ctx, procedure("P-1", "documents"), 2, answer); !errors.Is(err, engine.ErrNoCorrectionPending) {
		t.Errorf("a second AnswerCorrection() = %v, want ErrNoCorrectionPending", err)
	}
	third, thirdEvent := correction("P-1", "documents", []string{"subjects"}, nil)
	requestCorrectionOf(t, s, third, thirdEvent)

	want := start.Procedure
	want.CorrectionRequests = []domain.CorrectionRequest{third}
	wantRecorded(t, s, want, append(slices.Clone(start.Events), firstEvent, secondEvent, answer, thirdEvent), start.Values, start.Documents)
}

// answerBeforeAnotherRequest stores an answer that was made when the history
// had one request less: it does not answer the request it did not see.
func answerBeforeAnotherRequest(t *testing.T, s ProcedureStores) {
	ctx := t.Context()
	start := fullStart("P-1")
	add(t, s, start)
	first, firstEvent := correction("P-1", "documents", []string{"reason"}, nil)
	second, secondEvent := correction("P-1", "documents", nil, []string{"statement"})
	requestCorrectionOf(t, s, first, firstEvent)
	requestCorrectionOf(t, s, second, secondEvent)

	if err := s.Procedures.AnswerCorrection(ctx, procedure("P-1", "documents"), 1, answered("P-1", "documents", []string{"reason"}, nil)); !errors.Is(err, engine.ErrCorrectionsChanged) {
		t.Errorf("AnswerCorrection() read with one request = %v, want ErrCorrectionsChanged", err)
	}
	want := start.Procedure
	want.CorrectionRequests = []domain.CorrectionRequest{first, second}
	wantRecorded(t, s, want, append(slices.Clone(start.Events), firstEvent, secondEvent), start.Values, start.Documents)
}
