package postgres

import (
	"context"
	"errors"
	"fmt"

	"backend/src/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.DocumentRepository = (*DocumentRepository)(nil)

// DocumentRepository reads the documents that a ProcedureRepository records.
type DocumentRepository struct {
	pool *pgxpool.Pool
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{pool: pool}
}

func (r *DocumentRepository) List(ctx context.Context, procedureID string) ([]domain.Document, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+documentColumns+" FROM documents WHERE procedure_id = $1 ORDER BY number", procedureID)
	if err != nil {
		return nil, fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	documents, err := pgx.CollectRows(rows, scanDocument)
	if err != nil {
		return nil, fmt.Errorf("documents of %s: %w", procedureID, err)
	}
	return database.List(documents), nil
}

func (r *DocumentRepository) Get(ctx context.Context, procedureID, documentID string) (domain.Document, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+documentColumns+" FROM documents WHERE procedure_id = $1 AND id = $2", procedureID, documentID)
	if err != nil {
		return domain.Document{}, fmt.Errorf("read document %s: %w", documentID, err)
	}
	d, err := pgx.CollectExactlyOneRow(rows, scanDocument)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return domain.Document{}, fmt.Errorf("%w: %s", domain.ErrDocumentNotFound, documentID)
	case err != nil:
		return domain.Document{}, fmt.Errorf("read document %s: %w", documentID, err)
	}
	return d, nil
}

func (r *DocumentRepository) Digests(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT DISTINCT digest FROM documents")
	if err != nil {
		return nil, fmt.Errorf("digests of the documents: %w", err)
	}
	digests, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("digests of the documents: %w", err)
	}
	return database.List(digests), nil
}

func scanDocument(row pgx.CollectableRow) (domain.Document, error) {
	var d domain.Document
	err := row.Scan(&d.ID, &d.Procedure, &d.Requirement, &d.Step, &d.FileName, &d.MediaType, &d.Size, &d.Digest, &d.UploadedBy, &d.UploadedAt)
	d.UploadedAt = d.UploadedAt.UTC()
	return d, err
}
