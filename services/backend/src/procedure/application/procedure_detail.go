package application

import (
	"context"
	"fmt"

	"backend/src/procedure/domain"
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

// Detail fails, and the server answers 500, if the open step is not in the
// version the procedure started with.
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
	return input.ProcedureDetail{
		Procedure: p,
		Name:      content.Definition.Name,
		State:     state,
		Steps:     detail.WithValues(detail.WithFiles(detail.Steps(content), documents), values),
		CanAct:    canAct,
		Referral:  detail.ReferralOf(p, content, s.procedures.now()),
		History:   history,
	}, nil
}
