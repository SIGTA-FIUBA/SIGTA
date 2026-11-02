package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/domain/procedure"
	"backend/src/ports/input"
	"backend/src/ports/output"
)

var _ input.ProcedureTypes = (*ProcedureTypeService)(nil)

type ProcedureTypeService struct {
	versions output.ProcedureVersions
	areas    output.AreaCatalog
	now      func() time.Time
}

func NewProcedureTypeService(versions output.ProcedureVersions, areas output.AreaCatalog, now func() time.Time) *ProcedureTypeService {
	return &ProcedureTypeService{versions: versions, areas: areas, now: now}
}

func (s *ProcedureTypeService) Publish(ctx context.Context, p input.Publication) (procedure.Version, error) {
	if err := p.Authorship.Validate(); err != nil {
		return procedure.Version{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, err)
	}
	areas, err := s.areas.List(ctx)
	if err != nil {
		return procedure.Version{}, fmt.Errorf("list areas: %w", err)
	}
	if err := p.Definition.Validate(areas); err != nil {
		return procedure.Version{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, err)
	}
	encoded, err := procedure.NewContent(p.Definition, areas).Encode()
	if err != nil {
		return procedure.Version{}, err
	}

	latest, err := s.latest(ctx, p.ProcedureType)
	if err != nil {
		return procedure.Version{}, err
	}
	if unchanged, err := sameContent(latest, encoded); err != nil || unchanged {
		return latest, err
	}
	if latest.Number != p.Base {
		return procedure.Version{}, fmt.Errorf("procedure type %d: %w: edited from version %d, the latest is %d", p.ProcedureType, input.ErrOutdatedBase, p.Base, latest.Number)
	}

	v := procedure.Version{
		ProcedureType: p.ProcedureType,
		Number:        p.Base + 1,
		Digest:        procedure.Digest(encoded),
		Encoded:       encoded,
		Authorship:    p.Authorship,
		// PostgreSQL keeps microseconds, so the version returned here is the
		// one that is stored.
		PublishedAt: s.now().UTC().Truncate(time.Microsecond),
	}
	err = s.versions.Add(ctx, v)
	if errors.Is(err, output.ErrVersionExists) {
		return procedure.Version{}, fmt.Errorf("procedure type %d: %w: version %d was just published", p.ProcedureType, input.ErrOutdatedBase, v.Number)
	}
	if err != nil {
		return procedure.Version{}, fmt.Errorf("publish procedure type %d version %d: %w", p.ProcedureType, v.Number, err)
	}
	return v, nil
}

func (s *ProcedureTypeService) Version(ctx context.Context, procedureType int64, number int) (procedure.Version, error) {
	v, err := s.versions.Get(ctx, procedureType, number)
	if err == nil {
		err = check(v, procedureType, number)
	}
	if err != nil {
		return procedure.Version{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, number, err)
	}
	return v, nil
}

func (s *ProcedureTypeService) Catalog(ctx context.Context) ([]input.CatalogEntry, error) {
	versions, err := s.versions.LatestOfEach(ctx)
	if err != nil {
		return nil, fmt.Errorf("list latest versions: %w", err)
	}
	entries := make([]input.CatalogEntry, 0, len(versions))
	for _, version := range versions {
		if err := check(version, version.ProcedureType, version.Number); err != nil {
			return nil, fmt.Errorf("procedure type %d version %d: %w", version.ProcedureType, version.Number, err)
		}
		content, err := version.Content()
		if err != nil {
			return nil, fmt.Errorf("procedure type %d version %d: %w", version.ProcedureType, version.Number, err)
		}
		fields, documents := content.ApplicantRequirements()
		entries = append(entries, input.CatalogEntry{
			ProcedureType: version.ProcedureType,
			Version:       version.Number,
			Name:          content.Definition.Name,
			Description:   content.Definition.Description,
			Deadlines:     content.Definition.Deadlines,
			Admissibility: content.Definition.Admissibility,
			Fields:        fields,
			Documents:     documents,
		})
	}
	return entries, nil
}

// latest returns the zero Version if the procedure type has none, and fails
// if the latest one does not match its digest: publishing on top of it would
// hide the alteration.
func (s *ProcedureTypeService) latest(ctx context.Context, procedureType int64) (procedure.Version, error) {
	v, err := s.versions.Latest(ctx, procedureType)
	if errors.Is(err, output.ErrNoVersion) {
		return procedure.Version{}, nil
	}
	if err != nil {
		return procedure.Version{}, fmt.Errorf("procedure type %d: latest version: %w", procedureType, err)
	}
	if err := check(v, procedureType, v.Number); err != nil {
		return procedure.Version{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, v.Number, err)
	}
	return v, nil
}

// check fails if the store returned another row than the one asked for, since
// the digest covers the content and not where it is stored.
func check(v procedure.Version, procedureType int64, number int) error {
	if v.ProcedureType != procedureType || v.Number != number {
		return fmt.Errorf("the store returned version %d of procedure type %d", v.Number, v.ProcedureType)
	}
	return v.Verify()
}

// sameContent re-encodes the stored version instead of comparing bytes, so
// that a change in how Content is encoded does not publish a version that
// changes nothing.
func sameContent(v procedure.Version, encoded []byte) (bool, error) {
	if v.Number == 0 {
		return false, nil
	}
	c, err := v.Content()
	if err != nil {
		return false, fmt.Errorf("procedure type %d version %d: %w", v.ProcedureType, v.Number, err)
	}
	again, err := c.Encode()
	if err != nil {
		return false, err
	}
	return bytes.Equal(again, encoded), nil
}
