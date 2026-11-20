package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureRepository = (*ProcedureRepository)(nil)

// ProcedureRepository writes each change in one transaction. What it appends
// to a procedure, events and documents, takes the next number of that
// procedure while it holds the lock on the row of the procedure, so two
// changes to the same procedure take turns and those to different ones do
// not wait for each other.
type ProcedureRepository struct {
	pool *pgxpool.Pool
}

func NewProcedureRepository(pool *pgxpool.Pool) *ProcedureRepository {
	return &ProcedureRepository{pool: pool}
}

const (
	procedureColumns = "id, procedure_type_id, version_number, version_digest, applicant, open_step, referral_from, referral_to, referral_since, started_at, moved_at"
	eventColumns     = "procedure_id, step, action, done_by, done_at, observation, outcome, declared, send_back, next_step, referral_from, referral_to, external_file_number, returned, document_id, document_requirement, document_file_name, correction_fields, correction_documents"
	requestColumns   = "procedure_id, step, requested_by, requested_at, reason, fields, documents"
	valueColumns     = "procedure_id, field, step, text, items, entered_by, entered_at"
	documentColumns  = "id, procedure_id, requirement, step, file_name, media_type, size, digest, uploaded_by, uploaded_at"
)

func (s *ProcedureRepository) Add(ctx context.Context, start output.Start) error {
	p := start.Procedure
	err := database.InReadCommitted(ctx, s.pool, func(tx pgx.Tx) error {
		from, to, since := referralArgs(p.Referral)
		if _, err := tx.Exec(ctx, "INSERT INTO procedures ("+procedureColumns+") VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
			p.ID, p.ProcedureType, p.Version, p.VersionDigest, p.Applicant, database.Nullable(p.OpenStep), from, to, since, p.StartedAt, p.MovedAt); err != nil {
			if hasCode(err, uniqueViolation) {
				return fmt.Errorf("%w: %s: %w", output.ErrProcedureExists, p.ID, err)
			}
			return err
		}
		for i, v := range start.Values {
			if _, err := tx.Exec(ctx, "INSERT INTO procedure_field_values (number, "+valueColumns+") VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				i+1, p.ID, v.Field, v.Step, v.Text, database.Array(v.Items), v.EnteredBy, v.EnteredAt); err != nil {
				return err
			}
		}
		for _, e := range start.Events {
			if err := insertEvent(ctx, tx, p.ID, e); err != nil {
				return err
			}
		}
		for _, d := range start.Documents {
			if err := insertDocument(ctx, tx, p.ID, d); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("start procedure %s: %w", p.ID, err)
	}
	return nil
}

func (s *ProcedureRepository) Get(ctx context.Context, id string) (domain.Procedure, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+procedureColumns+" FROM procedures WHERE id = $1", id)
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("read procedure %s: %w", id, err)
	}
	p, err := pgx.CollectExactlyOneRow(rows, scanProcedure)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Procedure{}, fmt.Errorf("%w: %s", domain.ErrProcedureNotFound, id)
	case err != nil:
		return domain.Procedure{}, fmt.Errorf("read procedure %s: %w", id, err)
	}
	requests, err := s.correctionRequests(ctx, "WHERE procedure_id = $1", id)
	if err != nil {
		return domain.Procedure{}, fmt.Errorf("read procedure %s: %w", id, err)
	}
	p.CorrectionRequests = requests[id]
	return p, nil
}

// List orders by the bytes of the identifier, as Go compares strings, and not
// by the collation of the database.
func (s *ProcedureRepository) List(ctx context.Context) ([]domain.Procedure, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+procedureColumns+` FROM procedures ORDER BY id COLLATE "C"`)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	procedures, err := pgx.CollectRows(rows, scanProcedure)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	requests, err := s.correctionRequests(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	for i := range procedures {
		procedures[i].CorrectionRequests = requests[procedures[i].ID]
	}
	return procedures, nil
}

// correctionRequests returns the requests that match where, by procedure and
// in the order they were made.
func (s *ProcedureRepository) correctionRequests(ctx context.Context, where string, args ...any) (map[string][]domain.CorrectionRequest, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+requestColumns+" FROM procedure_correction_requests "+where+" ORDER BY procedure_id, number", args...)
	if err != nil {
		return nil, fmt.Errorf("correction requests: %w", err)
	}
	requests := map[string][]domain.CorrectionRequest{}
	var procedureID string
	var r domain.CorrectionRequest
	if _, err := pgx.ForEachRow(rows, []any{&procedureID, &r.Step, &r.By, &r.At, &r.Reason, &r.Fields, &r.Documents}, func() error {
		r.At = r.At.UTC()
		r.Fields, r.Documents = database.List(r.Fields), database.List(r.Documents)
		requests[procedureID] = append(requests[procedureID], r)
		r = domain.CorrectionRequest{}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("correction requests: %w", err)
	}
	return requests, nil
}

func (s *ProcedureRepository) Values(ctx context.Context, procedureID string) ([]domain.FieldValue, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+valueColumns+" FROM procedure_field_values WHERE procedure_id = $1 ORDER BY number", procedureID)
	if err != nil {
		return nil, fmt.Errorf("values of %s: %w", procedureID, err)
	}
	values, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.FieldValue, error) {
		var v domain.FieldValue
		err := row.Scan(&v.Procedure, &v.Field, &v.Step, &v.Text, &v.Items, &v.EnteredBy, &v.EnteredAt)
		v.Items = database.List(v.Items)
		v.EnteredAt = v.EnteredAt.UTC()
		return v, err
	})
	if err != nil {
		return nil, fmt.Errorf("values of %s: %w", procedureID, err)
	}
	return values, nil
}

// Advance checks the open step and the pending requests once it holds the
// lock on the row, so of two confirmations of the same step only the first
// one moves it, and a confirmation waits for a request being made at the same
// time and then sees it.
func (s *ProcedureRepository) Advance(ctx context.Context, p domain.Procedure, e domain.Event) error {
	err := database.InReadCommitted(ctx, s.pool, func(tx pgx.Tx) error {
		openStep, err := lockProcedure(ctx, tx, p.ID)
		if err != nil {
			return err
		}
		if openStep != e.Step {
			return engine.ErrStepNotOpen
		}
		var awaiting bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT FROM procedure_correction_requests WHERE procedure_id = $1)", p.ID).Scan(&awaiting); err != nil {
			return err
		}
		if awaiting {
			return engine.ErrAwaitingApplicant
		}
		from, to, since := referralArgs(p.Referral)
		if _, err := tx.Exec(ctx, `UPDATE procedures
			SET open_step = $2, referral_from = $3, referral_to = $4, referral_since = $5, moved_at = $6
			WHERE id = $1`,
			p.ID, database.Nullable(p.OpenStep), from, to, since, p.MovedAt); err != nil {
			return err
		}
		return insertEvent(ctx, tx, p.ID, e)
	})
	if err != nil {
		return fmt.Errorf("procedure %s: %w", p.ID, err)
	}
	return nil
}

func (s *ProcedureRepository) Attach(ctx context.Context, d domain.Document, e domain.Event) error {
	err := database.InReadCommitted(ctx, s.pool, func(tx pgx.Tx) error {
		openStep, err := lockProcedure(ctx, tx, d.Procedure)
		if err != nil {
			return err
		}
		if openStep != d.Step {
			return engine.ErrStepNotOpen
		}
		if err := insertDocument(ctx, tx, d.Procedure, d); err != nil {
			return err
		}
		return insertEvent(ctx, tx, d.Procedure, e)
	})
	if err != nil {
		return fmt.Errorf("attach %s to %s: %w", d.ID, d.Procedure, err)
	}
	return nil
}

func (s *ProcedureRepository) RequestCorrection(ctx context.Context, r domain.CorrectionRequest, e domain.Event) error {
	err := database.InReadCommitted(ctx, s.pool, func(tx pgx.Tx) error {
		openStep, err := lockProcedure(ctx, tx, e.ProcedureID)
		if err != nil {
			return err
		}
		if openStep != r.Step {
			return engine.ErrStepNotOpen
		}
		if _, err := tx.Exec(ctx, "INSERT INTO procedure_correction_requests (number, "+requestColumns+`)
			SELECT coalesce(max(number), 0) + 1, $1, $2, $3, $4, $5, $6, $7 FROM procedure_correction_requests WHERE procedure_id = $1`,
			e.ProcedureID, r.Step, r.By, r.At, r.Reason, database.Array(r.Fields), database.Array(r.Documents)); err != nil {
			return err
		}
		return insertEvent(ctx, tx, e.ProcedureID, e)
	})
	if err != nil {
		return fmt.Errorf("request a correction of %s: %w", e.ProcedureID, err)
	}
	return nil
}

func (s *ProcedureRepository) Events(ctx context.Context, procedureID string) ([]domain.Event, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+eventColumns+" FROM procedure_events WHERE procedure_id = $1 ORDER BY number", procedureID)
	if err != nil {
		return nil, fmt.Errorf("events of %s: %w", procedureID, err)
	}
	events, err := pgx.CollectRows(rows, scanEvent)
	if err != nil {
		return nil, fmt.Errorf("events of %s: %w", procedureID, err)
	}
	if len(events) == 0 {
		if _, err := s.Get(ctx, procedureID); err != nil {
			return nil, err
		}
	}
	return events, nil
}

// lockProcedure returns the open step of the procedure, empty if it is
// closed, and fails with domain.ErrProcedureNotFound if there is no such
// procedure.
func lockProcedure(ctx context.Context, tx pgx.Tx, id string) (string, error) {
	var openStep *string
	err := tx.QueryRow(ctx, "SELECT open_step FROM procedures WHERE id = $1 FOR NO KEY UPDATE", id).Scan(&openStep)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrProcedureNotFound
	}
	return database.ValueOf(openStep), err
}

// insertDocument fails with output.ErrDocumentExists if the identifier is
// taken. The caller holds the lock on the procedure, or the procedure is new.
func insertDocument(ctx context.Context, tx pgx.Tx, procedureID string, d domain.Document) error {
	_, err := tx.Exec(ctx, "INSERT INTO documents (number, "+documentColumns+`)
		SELECT coalesce(max(number), 0) + 1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10 FROM documents WHERE procedure_id = $2`,
		d.ID, procedureID, d.Requirement, d.Step, d.FileName, d.MediaType, d.Size, d.Digest, d.UploadedBy, d.UploadedAt)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "documents_pkey" {
		return fmt.Errorf("%w: %s: %w", output.ErrDocumentExists, d.ID, err)
	}
	return err
}

// insertEvent appends e to the history. The caller holds the lock on the
// procedure, or the procedure is new.
func insertEvent(ctx context.Context, tx pgx.Tx, procedureID string, e domain.Event) error {
	var from, to, fileNumber, documentID, requirement, fileName *string
	if e.Referral != nil {
		from, to, fileNumber = &e.Referral.From, &e.Referral.To, database.Nullable(e.Referral.ExternalFileNumber)
	}
	if e.Document != nil {
		documentID, requirement, fileName = &e.Document.ID, &e.Document.Requirement, &e.Document.FileName
	}
	var fields, documents []string
	if e.CorrectionRequest != nil {
		fields, documents = database.Array(e.CorrectionRequest.Fields), database.Array(e.CorrectionRequest.Documents)
	}
	_, err := tx.Exec(ctx, "INSERT INTO procedure_events (number, "+eventColumns+`)
		SELECT coalesce(max(number), 0) + 1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		FROM procedure_events WHERE procedure_id = $1`,
		procedureID, e.Step, string(e.Action), e.By, e.At, database.Nullable(e.Observation), database.Nullable(e.Outcome), e.Declared, e.SendBack,
		database.Nullable(e.Next), from, to, fileNumber, e.Returned, documentID, requirement, fileName, fields, documents)
	return err
}

func referralArgs(r *domain.Referral) (from, to *string, since *time.Time) {
	if r == nil {
		return nil, nil, nil
	}
	return &r.From, &r.To, &r.Since
}

func scanProcedure(row pgx.CollectableRow) (domain.Procedure, error) {
	var (
		p         domain.Procedure
		openStep  *string
		from, to  *string
		since     *time.Time
		startedAt time.Time
		movedAt   time.Time
	)
	if err := row.Scan(&p.ID, &p.ProcedureType, &p.Version, &p.VersionDigest, &p.Applicant, &openStep, &from, &to, &since, &startedAt, &movedAt); err != nil {
		return domain.Procedure{}, err
	}
	p.OpenStep = database.ValueOf(openStep)
	if from != nil {
		p.Referral = &domain.Referral{From: *from, To: database.ValueOf(to), Since: database.TimeOf(since)}
	}
	p.StartedAt, p.MovedAt = startedAt.UTC(), movedAt.UTC()
	return p, nil
}

func scanEvent(row pgx.CollectableRow) (domain.Event, error) {
	var (
		e                                 domain.Event
		action                            string
		observation, outcome, next        *string
		from, to, fileNumber              *string
		documentID, requirement, fileName *string
		fields, documents                 []string
	)
	if err := row.Scan(&e.ProcedureID, &e.Step, &action, &e.By, &e.At, &observation, &outcome, &e.Declared, &e.SendBack,
		&next, &from, &to, &fileNumber, &e.Returned, &documentID, &requirement, &fileName, &fields, &documents); err != nil {
		return domain.Event{}, err
	}
	e.Action = definition.Action(action)
	e.At = e.At.UTC()
	e.Observation, e.Outcome, e.Next = database.ValueOf(observation), database.ValueOf(outcome), database.ValueOf(next)
	if from != nil {
		e.Referral = &domain.ReferralRecord{From: *from, To: database.ValueOf(to), ExternalFileNumber: database.ValueOf(fileNumber)}
	}
	if documentID != nil {
		e.Document = &domain.DocumentRecord{ID: *documentID, Requirement: database.ValueOf(requirement), FileName: database.ValueOf(fileName)}
	}
	if fields != nil {
		e.CorrectionRequest = &domain.CorrectionRecord{Fields: database.List(fields), Documents: database.List(documents)}
	}
	return e, nil
}
