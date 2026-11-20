package application

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/detail"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.ProcedureDetails = (*ProcedureDetailService)(nil)

// ProcedureDetailService decides who sees a procedure and who acts on its open
// step as ProcedureService does, so both answer the same to the same person.
type ProcedureDetailService struct {
	procedures *ProcedureService
	documents  output.DocumentRepository
}

func NewProcedureDetailService(procedures *ProcedureService, documents output.DocumentRepository) *ProcedureDetailService {
	return &ProcedureDetailService{procedures: procedures, documents: documents}
}

// Detail shows the applicant only what detail.Applicant leaves, also when
// they are staff of an area it passes through. It fails, and the server
// answers 500, if the open step is not in the version the procedure started
// with.
func (s *ProcedureDetailService) Detail(ctx context.Context, userID, procedureID string) (input.ProcedureDetail, error) {
	p, content, err := s.procedures.visible(ctx, userID, procedureID)
	if err != nil {
		return input.ProcedureDetail{}, err
	}
	state, err := domain.StateOf(p, content.Definition)
	if err != nil {
		return input.ProcedureDetail{}, err
	}
	canAct := false
	if !p.Closed() {
		if canAct, err = s.procedures.canAct(ctx, userID, p, content, p.OpenStep); err != nil {
			return input.ProcedureDetail{}, err
		}
	}
	documents, err := s.documents.List(ctx, procedureID)
	if err != nil {
		return input.ProcedureDetail{}, fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	values, err := s.procedures.procedures.Values(ctx, procedureID)
	if err != nil {
		return input.ProcedureDetail{}, fmt.Errorf("values of %s: %w", procedureID, err)
	}
	history, err := s.procedures.procedures.Events(ctx, procedureID)
	if err != nil {
		return input.ProcedureDetail{}, fmt.Errorf("history of %s: %w", procedureID, err)
	}
	steps := detail.WithValues(detail.WithFiles(detail.Steps(content), documents), values)
	asApplicant := p.Applicant == userID
	if asApplicant {
		applicant := detail.ForApplicant(content, p)
		steps, history = applicant.Steps(steps), applicant.History(history)
	}
	return input.ProcedureDetail{
		Procedure:   p,
		Name:        content.Definition.Name,
		State:       state,
		Steps:       steps,
		CanAct:      canAct,
		AsApplicant: asApplicant,
		Referral:    detail.ReferralOf(p, content, s.procedures.now()),
		History:     history,
	}, nil
}

// Mine leaves out the closed procedures and those of others, also for staff:
// what goes through their areas is in the inbox. It fails, as the detail
// does, if the open step of a procedure is not in its version.
func (s *ProcedureDetailService) Mine(ctx context.Context, userID string) ([]input.MyProcedure, error) {
	all, err := s.procedures.procedures.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	mine := []input.MyProcedure{}
	for _, p := range all {
		if p.Applicant != userID || p.Closed() {
			continue
		}
		content, err := s.procedures.content(ctx, p)
		if err != nil {
			return nil, err
		}
		state, err := domain.StateOf(p, content.Definition)
		if err != nil {
			return nil, err
		}
		// StateOf already found the open step in the definition.
		step := content.Definition.Steps[slices.IndexFunc(content.Definition.Steps, func(st definition.Step) bool { return st.ID == p.OpenStep })]
		mine = append(mine, input.MyProcedure{Procedure: p, Name: content.Definition.Name, StepName: step.Name, State: state})
	}
	slices.SortFunc(mine, func(a, b input.MyProcedure) int {
		return cmp.Or(b.Procedure.MovedAt.Compare(a.Procedure.MovedAt), cmp.Compare(a.Procedure.ID, b.Procedure.ID))
	})
	return mine, nil
}
