package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var _ input.ProcedureTypes = (*ProcedureTypeService)(nil)

type ProcedureTypeService struct {
	versions output.ProcedureVersions
	areas    authinput.Areas
	now      func() time.Time
}

func NewProcedureTypeService(versions output.ProcedureVersions, areas authinput.Areas, now func() time.Time) *ProcedureTypeService {
	return &ProcedureTypeService{versions: versions, areas: areas, now: now}
}

func (s *ProcedureTypeService) Publish(ctx context.Context, p input.Publication) (domain.Version, error) {
	if err := p.Authorship.Validate(); err != nil {
		return domain.Version{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, err)
	}
	areas, err := s.stepAreas(ctx)
	if err != nil {
		return domain.Version{}, err
	}
	if err := p.Definition.Validate(areas); err != nil {
		return domain.Version{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, err)
	}
	encoded, err := domain.NewContent(p.Definition, areas).Encode()
	if err != nil {
		return domain.Version{}, err
	}

	latest, err := s.latest(ctx, p.ProcedureType)
	if err != nil {
		return domain.Version{}, err
	}
	unchanged, err := sameContent(latest, encoded)
	if err != nil {
		return domain.Version{}, err
	}
	if unchanged {
		return latest, nil
	}
	if latest.Number != p.Base {
		return domain.Version{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, &input.OutdatedBaseError{Base: p.Base, Latest: latest.Number})
	}

	v := domain.Version{
		ProcedureType: p.ProcedureType,
		Number:        p.Base + 1,
		Digest:        domain.Digest(encoded),
		Encoded:       encoded,
		Authorship:    p.Authorship,
		// PostgreSQL keeps microseconds, so the version returned here is the
		// one that is stored.
		PublishedAt: s.now().UTC().Truncate(time.Microsecond),
	}
	err = s.versions.Add(ctx, v)
	if errors.Is(err, output.ErrVersionExists) {
		return domain.Version{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, &input.OutdatedBaseError{Base: p.Base, Latest: v.Number})
	}
	if err != nil {
		return domain.Version{}, fmt.Errorf("publish procedure type %d version %d: %w", p.ProcedureType, v.Number, err)
	}
	return v, nil
}

// stepAreas lists the staff areas of auth, which a version copies as they are
// now.
func (s *ProcedureTypeService) stepAreas(ctx context.Context) ([]domain.Area, error) {
	staff, err := s.areas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	areas := make([]domain.Area, 0, len(staff))
	for _, a := range staff {
		areas = append(areas, domain.Area{ID: a.ID, Name: a.Name})
	}
	return areas, nil
}

func (s *ProcedureTypeService) Version(ctx context.Context, procedureType int64, number int) (domain.Version, error) {
	v, err := s.versions.Get(ctx, procedureType, number)
	if err != nil {
		return domain.Version{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, number, err)
	}
	if err := check(v, procedureType, number); err != nil {
		return domain.Version{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, number, err)
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
func (s *ProcedureTypeService) latest(ctx context.Context, procedureType int64) (domain.Version, error) {
	v, err := s.versions.Latest(ctx, procedureType)
	if errors.Is(err, output.ErrNoVersion) {
		return domain.Version{}, nil
	}
	if err != nil {
		return domain.Version{}, fmt.Errorf("procedure type %d: latest version: %w", procedureType, err)
	}
	if err := check(v, procedureType, v.Number); err != nil {
		return domain.Version{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, v.Number, err)
	}
	return v, nil
}

// check fails if the store returned another row than the one asked for, since
// the digest covers the content and not where it is stored.
func check(v domain.Version, procedureType int64, number int) error {
	if v.ProcedureType != procedureType || v.Number != number {
		return fmt.Errorf("%w: version %d of procedure type %d", output.ErrWrongVersion, v.Number, v.ProcedureType)
	}
	return v.Verify()
}

// sameContent re-encodes the stored version instead of comparing bytes, so
// that a change in how Content is encoded does not publish a version that
// changes nothing.
func sameContent(v domain.Version, encoded []byte) (bool, error) {
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
