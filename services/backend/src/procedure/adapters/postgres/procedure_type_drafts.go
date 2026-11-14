package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/procedure/domain/draft"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureTypeDraftRepository = (*ProcedureTypeDraftRepository)(nil)

// ProcedureTypeDraftRepository changes a draft with a single statement that checks its
// revision, so it needs no transaction and behaves the same at READ COMMITTED
// and at SERIALIZABLE.
type ProcedureTypeDraftRepository struct {
	pool *pgxpool.Pool
}

func NewProcedureTypeDraftRepository(pool *pgxpool.Pool) *ProcedureTypeDraftRepository {
	return &ProcedureTypeDraftRepository{pool: pool}
}

const draftColumns = "procedure_type_id, base, definition, status, revision, edited_by, edited_at, proposed_by, proposed_at, rejected_by, rejected_at, rejection_reason"

func (s *ProcedureTypeDraftRepository) Get(ctx context.Context, procedureType int64) (draft.Draft, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+draftColumns+" FROM procedure_type_drafts WHERE procedure_type_id = $1", procedureType)
	if err != nil {
		return draft.Draft{}, fmt.Errorf("read the draft of procedure type %d: %w", procedureType, err)
	}
	d, err := pgx.CollectExactlyOneRow(rows, scanDraft)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return draft.Draft{}, fmt.Errorf("%w: procedure type %d", output.ErrDraftNotFound, procedureType)
	case err != nil:
		return draft.Draft{}, fmt.Errorf("read the draft of procedure type %d: %w", procedureType, err)
	}
	return d, nil
}

func (s *ProcedureTypeDraftRepository) List(ctx context.Context) ([]draft.Draft, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+draftColumns+" FROM procedure_type_drafts ORDER BY procedure_type_id")
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}
	drafts, err := pgx.CollectRows(rows, scanDraft)
	if err != nil {
		return nil, fmt.Errorf("list drafts: %w", err)
	}
	return drafts, nil
}

func (s *ProcedureTypeDraftRepository) Add(ctx context.Context, d draft.Draft) error {
	args, err := draftArgs(d)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, "INSERT INTO procedure_type_drafts ("+draftColumns+") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)", args...)
	if hasCode(err, uniqueViolation) {
		return fmt.Errorf("%w: procedure type %d: %w", output.ErrDraftExists, d.ProcedureTypeID, err)
	}
	if err != nil {
		return fmt.Errorf("add the draft of procedure type %d: %w", d.ProcedureTypeID, err)
	}
	return nil
}

func (s *ProcedureTypeDraftRepository) Update(ctx context.Context, d draft.Draft) error {
	args, err := draftArgs(d)
	if err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx, `UPDATE procedure_type_drafts
		SET definition = $3, status = $4, revision = $5, edited_by = $6, edited_at = $7,
			proposed_by = $8, proposed_at = $9, rejected_by = $10, rejected_at = $11, rejection_reason = $12
		WHERE procedure_type_id = $1 AND base = $2 AND revision = $5 - 1`, args...)
	if err != nil {
		return fmt.Errorf("update the draft of procedure type %d: %w", d.ProcedureTypeID, err)
	}
	return s.changed(ctx, tag.RowsAffected(), d.ProcedureTypeID)
}

func (s *ProcedureTypeDraftRepository) Remove(ctx context.Context, procedureType int64, revision int) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM procedure_type_drafts WHERE procedure_type_id = $1 AND revision = $2", procedureType, revision)
	if err != nil {
		return fmt.Errorf("remove the draft of procedure type %d: %w", procedureType, err)
	}
	return s.changed(ctx, tag.RowsAffected(), procedureType)
}

// changed explains a statement that changed no row: either there is no draft
// or it is at another revision.
func (s *ProcedureTypeDraftRepository) changed(ctx context.Context, rows int64, procedureType int64) error {
	if rows > 0 {
		return nil
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, "SELECT EXISTS (SELECT FROM procedure_type_drafts WHERE procedure_type_id = $1)", procedureType).Scan(&exists); err != nil {
		return fmt.Errorf("read the draft of procedure type %d: %w", procedureType, err)
	}
	if !exists {
		return fmt.Errorf("%w: procedure type %d", output.ErrDraftNotFound, procedureType)
	}
	return fmt.Errorf("%w: procedure type %d", output.ErrDraftChanged, procedureType)
}

func draftArgs(d draft.Draft) ([]any, error) {
	def, err := json.Marshal(d.Definition)
	if err != nil {
		return nil, fmt.Errorf("encode the draft of procedure type %d: %w", d.ProcedureTypeID, err)
	}
	return []any{
		d.ProcedureTypeID, d.Base, def, string(d.Status), d.Revision, d.EditedBy, d.EditedAt,
		nullable(d.ProposedBy), nullableTime(d.ProposedAt),
		nullable(d.Rejection.By), nullableTime(d.Rejection.At), nullable(d.Rejection.Reason),
	}, nil
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func scanDraft(row pgx.CollectableRow) (draft.Draft, error) {
	var (
		d                              draft.Draft
		def                            []byte
		status                         string
		proposedBy, rejectedBy, reason *string
		proposedAt, rejectedAt         *time.Time
	)
	err := row.Scan(&d.ProcedureTypeID, &d.Base, &def, &status, &d.Revision, &d.EditedBy, &d.EditedAt,
		&proposedBy, &proposedAt, &rejectedBy, &rejectedAt, &reason)
	if err != nil {
		return draft.Draft{}, err
	}
	if err := json.Unmarshal(def, &d.Definition); err != nil {
		return draft.Draft{}, fmt.Errorf("decode the draft of procedure type %d: %w", d.ProcedureTypeID, err)
	}
	d.Status = draft.Status(status)
	d.EditedAt = d.EditedAt.UTC()
	d.ProposedBy, d.ProposedAt = valueOf(proposedBy), timeOf(proposedAt)
	d.Rejection = draft.Rejection{By: valueOf(rejectedBy), At: timeOf(rejectedAt), Reason: valueOf(reason)}
	return d, nil
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timeOf(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.UTC()
}
