package application

import (
	"context"
	"fmt"
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/submission"
)

const actionAnswer = "answer a correction of"

// EnterValue records a value the applicant enters in answer to a pending
// correction, on the step of theirs that asks for the field. Only the
// applicant enters it, and only what a pending correction asks for.
func (s *ProcedureService) EnterValue(ctx context.Context, userID, procedureID, fieldID string, v submission.Value) (domain.FieldValue, error) {
	p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return domain.FieldValue{}, err
	}
	if p.Applicant != userID {
		return domain.FieldValue{}, denied(userID, actionAnswer+" "+p.ID)
	}
	step, err := engine.CorrectionField(p, content.Definition, fieldID)
	if err != nil {
		return domain.FieldValue{}, fmt.Errorf("procedure %s field %q: %w", p.ID, fieldID, err)
	}
	i := slices.IndexFunc(content.Definition.Fields, func(f definition.Field) bool { return f.ID == fieldID })
	if i < 0 {
		return domain.FieldValue{}, fmt.Errorf("step %q asks for %q, which is not in the definition", step.ID, fieldID)
	}
	clean, ok := submission.CheckValue(content.Definition.Fields[i], v)
	if !ok {
		return domain.FieldValue{}, fmt.Errorf("procedure %s field %q: %w", p.ID, fieldID, submission.ErrInvalidValue)
	}
	at := s.now()
	value := domain.FieldValue{Procedure: p.ID, Field: fieldID, Step: step.ID, Text: clean.Text, Items: clean.Items, EnteredBy: userID, EnteredAt: at}
	event := domain.Event{ProcedureID: p.ID, Step: step.ID, Action: definition.ActionEnterData, By: userID, At: at, Value: &domain.ValueRecord{Field: fieldID}}
	if err := s.procedures.EnterValue(ctx, value, event); err != nil {
		return domain.FieldValue{}, fmt.Errorf("enter %q on procedure %s: %w", fieldID, p.ID, err)
	}
	return value, nil
}

// AnswerCorrection sends what the pending corrections asked for, once the
// applicant loaded all of it, and takes the procedure back to staff.
func (s *ProcedureService) AnswerCorrection(ctx context.Context, userID, procedureID string) (domain.Event, error) {
	p, content, err := s.visible(ctx, userID, procedureID)
	if err != nil {
		return domain.Event{}, err
	}
	if p.Applicant != userID {
		return domain.Event{}, denied(userID, actionAnswer+" "+p.ID)
	}
	history, err := s.procedures.Events(ctx, p.ID)
	if err != nil {
		return domain.Event{}, fmt.Errorf("history of %s: %w", p.ID, err)
	}
	// Read again after the history, so that a request that came in between
	// is pending here too: what is stored is then never older than what the
	// history counts, and a request after it is refused when storing.
	if p, err = s.procedures.Get(ctx, p.ID); err != nil {
		return domain.Event{}, fmt.Errorf("procedure %s: %w", procedureID, err)
	}
	answered, event, err := engine.AnswerCorrection(p, content.Definition, history, userID, s.now())
	if err != nil {
		return domain.Event{}, fmt.Errorf("procedure %s: %w", p.ID, err)
	}
	if err := s.procedures.AnswerCorrection(ctx, answered, engine.CorrectionRequestsIn(history), event); err != nil {
		return domain.Event{}, fmt.Errorf("answer the corrections of procedure %s: %w", p.ID, err)
	}
	return event, nil
}
