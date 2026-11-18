// Package postgres stores the procedure module in PostgreSQL, with the tables
// of src/migrations and the privileges of the application role.
package postgres

import (
	"cmp"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/output"
)

const (
	uniqueViolation      = "23505"
	checkViolation       = "23514"
	serializationFailure = "40001"

	// approverIsNotProposer is the constraint of src/migrations that compares
	// the authors with lower(), which depends on the locale of the database.
	approverIsNotProposer = "approver_is_not_proposer"

	// digestIsOfContent is the constraint of src/migrations that recomputes
	// the digest of the content.
	digestIsOfContent = "digest_is_of_content"

	// attempts bounds how many times Add tries a version whose number is
	// still free after a serialization failure.
	attempts = 3
)

var _ output.ProcedureVersions = (*ProcedureVersions)(nil)

// ProcedureVersions never decodes the content: it reads and writes the bytes
// the digest covers. It works at READ COMMITTED, the default of PostgreSQL,
// and at SERIALIZABLE, where its contract tests also run.
type ProcedureVersions struct {
	pool *pgxpool.Pool
	// beforeCommit, if set, runs in each attempt of Add after the insert, so
	// a test can publish a rival version at that point.
	beforeCommit func(ctx context.Context)
}

func NewProcedureVersions(pool *pgxpool.Pool) *ProcedureVersions {
	return &ProcedureVersions{pool: pool}
}

const versionColumns = "procedure_type_id, number, digest, content, proposed_by, approved_by, published_at"

func (s *ProcedureVersions) Latest(ctx context.Context, procedureType int64) (version.ProcedureTypeVersion, error) {
	return s.one(ctx, fmt.Sprintf("latest version of procedure type %d", procedureType),
		"SELECT "+versionColumns+" FROM current_procedure_type_versions WHERE procedure_type_id = $1", procedureType)
}

func (s *ProcedureVersions) LatestOfEach(ctx context.Context) ([]version.ProcedureTypeVersion, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+versionColumns+" FROM current_procedure_type_versions ORDER BY procedure_type_id")
	if err != nil {
		return nil, fmt.Errorf("latest versions: %w", err)
	}
	latest, err := pgx.CollectRows(rows, scanVersion)
	if err != nil {
		return nil, fmt.Errorf("latest versions: %w", err)
	}
	return latest, nil
}

func (s *ProcedureVersions) Get(ctx context.Context, procedureType int64, number int) (version.ProcedureTypeVersion, error) {
	return s.one(ctx, fmt.Sprintf("version %d of procedure type %d", number, procedureType),
		"SELECT "+versionColumns+" FROM procedure_type_versions WHERE procedure_type_id = $1 AND number = $2", procedureType, number)
}

// Add inserts the version only if the previous number is the latest one, so
// the numbers have no gaps: the database has no constraint for that. A
// concurrent publication of the same number shows up in three ways, all of
// them ErrVersionExists: the primary key is violated (23505), the condition
// finds that the rival already committed and inserts nothing, or a
// serializable transaction fails (40001). In the last two the version is read
// again, since nothing else says why it failed. At SERIALIZABLE a version of
// another procedure type can also cause a serialization failure, since the
// predicate locks cover pages and not procedure types: if the number is
// still free, Add tries again, up to attempts times, and then returns the
// failure as it is, with its 40001 and without ErrVersionExists.
func (s *ProcedureVersions) Add(ctx context.Context, v version.ProcedureTypeVersion) error {
	for attempt := 1; ; attempt++ {
		inserted, err := s.insert(ctx, v)
		switch {
		case inserted:
			return nil
		case err != nil && !hasCode(err, serializationFailure):
			return insertError(v, err)
		}

		latest, readErr := s.Latest(ctx, v.ProcedureTypeID)
		if readErr != nil && !errors.Is(readErr, output.ErrNoVersion) {
			return fmt.Errorf("add version %d of procedure type %d: %w", v.Number, v.ProcedureTypeID, errors.Join(err, readErr))
		}
		next := latest.Number + 1
		switch {
		case v.Number >= 1 && v.Number < next:
			return versionError(output.ErrVersionExists, v, err)
		case v.Number != next:
			return versionError(output.ErrVersionGap, v, err)
		case attempt < attempts:
			continue
		}
		return fmt.Errorf("add version %d of procedure type %d, after %d attempts: %w", v.Number, v.ProcedureTypeID, attempts, cmp.Or(err, errNotInserted))
	}
}

// errNotInserted is the cause of a failed Add when no attempt inserted the
// version and none failed.
var errNotInserted = errors.New("the version was not inserted")

// insertError explains an insert that failed other than by serialization.
func insertError(v version.ProcedureTypeVersion, err error) error {
	switch {
	case violatesCheck(err, approverIsNotProposer):
		// Authorship.Validate compares with strings.EqualFold, so this is a
		// name that the locale of the database folds otherwise.
		return fmt.Errorf("%w: the database takes approver %q for proposer %q: %w", version.ErrInvalidAuthorship, v.Authorship.ApprovedBy, v.Authorship.ProposedBy, err)
	case violatesCheck(err, digestIsOfContent):
		return fmt.Errorf("%w: version %d of procedure type %d: %w", version.ErrAlteredContent, v.Number, v.ProcedureTypeID, err)
	case hasCode(err, uniqueViolation):
		return versionError(output.ErrVersionExists, v, err)
	}
	return fmt.Errorf("add version %d of procedure type %d: %w", v.Number, v.ProcedureTypeID, err)
}

func versionError(sentinel error, v version.ProcedureTypeVersion, cause error) error {
	err := fmt.Errorf("%w: version %d of procedure type %d", sentinel, v.Number, v.ProcedureTypeID)
	if cause == nil {
		return err
	}
	return fmt.Errorf("%w: %w", err, cause)
}

// insert also registers the procedure type with its first version, since
// versions reference it and nothing else creates procedure types yet.
func (s *ProcedureVersions) insert(ctx context.Context, v version.ProcedureTypeVersion) (bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if v.Number == 1 {
		if _, err := tx.Exec(ctx, "INSERT INTO procedure_types (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", v.ProcedureTypeID); err != nil {
			return false, err
		}
	}
	tag, err := tx.Exec(ctx, `INSERT INTO procedure_type_versions (`+versionColumns+`)
		SELECT $1, $2, $3, $4, $5, $6, $7
		WHERE (SELECT coalesce(max(number), 0) FROM procedure_type_versions WHERE procedure_type_id = $1) = $2 - 1`,
		v.ProcedureTypeID, v.Number, v.Digest, v.EncodedContent, v.Authorship.ProposedBy, v.Authorship.ApprovedBy, v.PublishedAt)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if s.beforeCommit != nil {
		s.beforeCommit(ctx)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return true, nil
}

func hasCode(err error, code string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == code
}

func violatesCheck(err error, constraint string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == checkViolation && pgErr.ConstraintName == constraint
}

// one reads the version that sql selects, described by what.
func (s *ProcedureVersions) one(ctx context.Context, what, sql string, args ...any) (version.ProcedureTypeVersion, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("read %s: %w", what, err)
	}
	v, err := pgx.CollectExactlyOneRow(rows, scanVersion)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return version.ProcedureTypeVersion{}, fmt.Errorf("%w: %s", output.ErrNoVersion, what)
	case err != nil:
		return version.ProcedureTypeVersion{}, fmt.Errorf("read %s: %w", what, err)
	}
	return v, nil
}

func scanVersion(row pgx.CollectableRow) (version.ProcedureTypeVersion, error) {
	var v version.ProcedureTypeVersion
	if err := row.Scan(&v.ProcedureTypeID, &v.Number, &v.Digest, &v.EncodedContent, &v.Authorship.ProposedBy, &v.Authorship.ApprovedBy, &v.PublishedAt); err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	v.PublishedAt = v.PublishedAt.UTC()
	return v, nil
}
