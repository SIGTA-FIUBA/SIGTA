package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/domain/catalog"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
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

func (s *ProcedureTypeService) Publish(ctx context.Context, p input.Publication) (version.ProcedureTypeVersion, error) {
	if err := p.Authorship.Validate(); err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, err)
	}
	areas, err := s.currentStaffAreas(ctx)
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	if err := p.Definition.Validate(areas); err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, err)
	}
	encoded, err := version.NewContent(p.Definition, areas).Encode()
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}

	latest, err := s.latest(ctx, p.ProcedureType)
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	unchanged, err := sameContent(latest, encoded)
	if err != nil {
		return version.ProcedureTypeVersion{}, err
	}
	if unchanged {
		return latest, nil
	}
	if latest.Number != p.Base {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, &input.OutdatedBaseError{Base: p.Base, Latest: latest.Number})
	}

	v := version.ProcedureTypeVersion{
		ProcedureTypeID: p.ProcedureType,
		Number:          p.Base + 1,
		Digest:          version.Digest(encoded),
		EncodedContent:  encoded,
		Authorship:      p.Authorship,
		// PostgreSQL keeps microseconds, so the version returned here is the
		// one that is stored.
		PublishedAt: s.now().UTC().Truncate(time.Microsecond),
	}
	err = s.versions.Add(ctx, v)
	if errors.Is(err, output.ErrVersionExists) {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: %w", p.ProcedureType, &input.OutdatedBaseError{Base: p.Base, Latest: v.Number})
	}
	if err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("publish procedure type %d version %d: %w", p.ProcedureType, v.Number, err)
	}
	return v, nil
}

func (s *ProcedureTypeService) currentStaffAreas(ctx context.Context) ([]definition.Area, error) {
	staff, err := s.areas.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	areas := make([]definition.Area, 0, len(staff))
	for _, a := range staff {
		area := definition.Area{ID: a.ID, Name: a.Name}
		for _, p := range a.Permissions {
			area.Permissions = append(area.Permissions, definition.Permission{ID: p.ID, Name: p.Name})
		}
		areas = append(areas, area)
	}
	return areas, nil
}

func (s *ProcedureTypeService) Version(ctx context.Context, procedureType int64, number int) (version.ProcedureTypeVersion, error) {
	v, err := s.versions.Get(ctx, procedureType, number)
	if err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, number, err)
	}
	if err := check(v, procedureType, number); err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, number, err)
	}
	return v, nil
}

func (s *ProcedureTypeService) Catalog(ctx context.Context) ([]input.CatalogEntry, error) {
	versions, err := s.versions.LatestOfEach(ctx)
	if err != nil {
		return nil, fmt.Errorf("list latest versions: %w", err)
	}
	entries := make([]input.CatalogEntry, 0, len(versions))
	for _, v := range versions {
		if err := check(v, v.ProcedureTypeID, v.Number); err != nil {
			return nil, fmt.Errorf("procedure type %d version %d: %w", v.ProcedureTypeID, v.Number, err)
		}
		content, err := v.DecodeContent()
		if err != nil {
			return nil, fmt.Errorf("procedure type %d version %d: %w", v.ProcedureTypeID, v.Number, err)
		}
		fields, documents := catalog.ApplicantRequirements(content)
		entries = append(entries, input.CatalogEntry{
			ProcedureType: v.ProcedureTypeID,
			Version:       v.Number,
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

// latest returns the zero ProcedureTypeVersion if the procedure type has none,
// and fails if the latest one does not match its digest: publishing on top of
// it would hide the alteration.
func (s *ProcedureTypeService) latest(ctx context.Context, procedureType int64) (version.ProcedureTypeVersion, error) {
	v, err := s.versions.Latest(ctx, procedureType)
	if errors.Is(err, output.ErrNoVersion) {
		return version.ProcedureTypeVersion{}, nil
	}
	if err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d: latest version: %w", procedureType, err)
	}
	if err := check(v, procedureType, v.Number); err != nil {
		return version.ProcedureTypeVersion{}, fmt.Errorf("procedure type %d version %d: %w", procedureType, v.Number, err)
	}
	return v, nil
}

// check fails if the store returned another row than the one asked for, since
// the digest covers the content and not where it is stored.
func check(v version.ProcedureTypeVersion, procedureType int64, number int) error {
	if v.ProcedureTypeID != procedureType || v.Number != number {
		return fmt.Errorf("%w: version %d of procedure type %d", output.ErrWrongVersion, v.Number, v.ProcedureTypeID)
	}
	return v.VerifyDigest()
}

// sameContent re-encodes the stored version instead of comparing bytes, so
// that a change in how Content is encoded does not publish a version that
// changes nothing.
func sameContent(v version.ProcedureTypeVersion, encoded []byte) (bool, error) {
	if v.Number == 0 {
		return false, nil
	}
	c, err := v.DecodeContent()
	if err != nil {
		return false, fmt.Errorf("procedure type %d version %d: %w", v.ProcedureTypeID, v.Number, err)
	}
	again, err := c.Encode()
	if err != nil {
		return false, err
	}
	return bytes.Equal(again, encoded), nil
}
