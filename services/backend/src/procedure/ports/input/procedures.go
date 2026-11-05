package input

import (
	"context"

	"backend/src/procedure/domain"
)

// ProcedureView is a procedure as the actor who opened it sees it.
type ProcedureView struct {
	Procedure domain.Procedure
	// Actions are the steps the actor is offered to act on.
	Actions []string
}

// Procedures checks on the server who can see and act on each procedure.
// Every refusal for lack of permission is recorded as a denied attempt.
type Procedures interface {
	// Register fails with domain.ErrUnknownUserKind for a kind outside the
	// three.
	Register(ctx context.Context, userID string, kind domain.UserKind) error
	// Grant fails with domain.ErrInvalidPermission if the person is not staff,
	// the profile is unknown or the procedure type is not published.
	Grant(ctx context.Context, p domain.Permission) error
	// Start opens a procedure on the first step of the latest version. It
	// fails with domain.ErrAccessDenied for anyone but an applicant.
	Start(ctx context.Context, userID string, procedureType int64) (domain.Procedure, error)
	// Open fails with domain.ErrProcedureNotFound both when the procedure does
	// not exist and when the actor cannot see it.
	Open(ctx context.Context, userID, procedureID string) (ProcedureView, error)
	List(ctx context.Context, userID string) ([]domain.Procedure, error)
	// Act fails with domain.ErrAccessDenied if the step is not the open one or
	// the actor cannot act on it, and leaves the procedure as it was.
	Act(ctx context.Context, userID, procedureID, stepID string) error
}
