package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/src/adapters/memory"
	"backend/src/domain/procedure"
	"backend/src/ports/input"
	"backend/src/ports/output"
)

var (
	applicant   = procedure.Area{ID: "applicant", Name: "Titular", Kind: procedure.AreaApplicant}
	publishedAt = time.Date(2026, 11, 1, 12, 0, 0, 123456789, time.UTC)
	authors     = procedure.Authorship{ProposedBy: "ana", ApprovedBy: "beto"}
	errStore    = errors.New("store unavailable")
)

func testDefinition() procedure.Definition {
	return procedure.Definition{
		Name:        "Trámite de prueba",
		Description: "Un trámite para las pruebas",
		Fields:      []procedure.Field{{ID: "national_id", Name: "DNI", Type: procedure.FieldText}},
		Steps: []procedure.Step{
			{ID: "request", Name: "Solicitud", Area: "applicant", Action: procedure.ActionEnterData, Fields: []string{"national_id"}},
		},
	}
}

func publication(procedureType int64, base int, d procedure.Definition) input.Publication {
	return input.Publication{ProcedureType: procedureType, Base: base, Definition: d, Authorship: authors}
}

func newService(versions output.ProcedureVersions, areas ...procedure.Area) *ProcedureTypeService {
	return NewProcedureTypeService(versions, memory.NewAreaCatalog(areas...), func() time.Time { return publishedAt })
}

// fakeVersions wraps the memory adapter so that a test can make it fail or
// lose a race.
type fakeVersions struct {
	*memory.ProcedureVersions
	latestErr error
	addErr    error
	stale     bool
	wrongRow  bool
}

func (f *fakeVersions) Get(ctx context.Context, procedureType int64, number int) (procedure.Version, error) {
	if f.wrongRow {
		return f.ProcedureVersions.Get(ctx, procedureType+1, number)
	}
	return f.ProcedureVersions.Get(ctx, procedureType, number)
}

func (f *fakeVersions) Latest(ctx context.Context, procedureType int64) (procedure.Version, error) {
	if f.latestErr != nil {
		return procedure.Version{}, f.latestErr
	}
	if f.stale {
		f.stale = false
		return procedure.Version{}, output.ErrNoVersion
	}
	return f.ProcedureVersions.Latest(ctx, procedureType)
}

func (f *fakeVersions) Add(ctx context.Context, v procedure.Version) error {
	if f.addErr != nil {
		return f.addErr
	}
	return f.ProcedureVersions.Add(ctx, v)
}

type failingCatalog struct{}

func (failingCatalog) List(context.Context) ([]procedure.Area, error) { return nil, errStore }

func TestPublish(t *testing.T) {
	ctx := context.Background()
	service := newService(memory.NewProcedureVersions(), applicant)
	publish := func(p input.Publication) int {
		t.Helper()
		v, err := service.Publish(ctx, p)
		if err != nil {
			t.Fatalf("Publish() = %v", err)
		}
		return v.Number
	}

	if got := publish(publication(1, 0, testDefinition())); got != 1 {
		t.Errorf("first publication: version %d, want 1", got)
	}
	if got := publish(publication(1, 1, testDefinition())); got != 1 {
		t.Errorf("publication without changes: version %d, want 1", got)
	}
	again := publication(1, 0, testDefinition())
	again.Authorship = procedure.Authorship{ProposedBy: "carla", ApprovedBy: "dario"}
	stored, _ := service.Version(ctx, 1, 1)
	if v, err := service.Publish(ctx, again); err != nil || !reflect.DeepEqual(v, stored) {
		t.Errorf("Publish() without changes by others = %+v, %v, want version 1 as stored", v, err)
	}

	changed := testDefinition()
	changed.Steps[0].Name = "Solicitud de equivalencia"
	if got := publish(publication(1, 1, changed)); got != 2 {
		t.Errorf("changed step: version %d, want 2", got)
	}

	changed.Fields[0].Name = "Documento"
	if got := publish(publication(1, 2, changed)); got != 3 {
		t.Errorf("changed field: version %d, want 3", got)
	}

	if got := publish(publication(2, 0, testDefinition())); got != 1 {
		t.Errorf("another procedure type: version %d, want 1", got)
	}

	v, err := service.Version(ctx, 1, 3)
	if err != nil {
		t.Fatalf("Version() = %v", err)
	}
	if want := publishedAt.Truncate(time.Microsecond); v.Authorship != authors || !v.PublishedAt.Equal(want) {
		t.Errorf("Version() authorship %v at %v, want %v at %v", v.Authorship, v.PublishedAt, authors, want)
	}
	if c, err := v.Content(); err != nil || c.Definition.Fields[0].Name != "Documento" {
		t.Errorf("Content() = %v, %v, want the third definition", c, err)
	}
}

func TestPublishFromAnOutdatedBase(t *testing.T) {
	ctx := context.Background()
	service := newService(memory.NewProcedureVersions(), applicant)
	if _, err := service.Publish(ctx, publication(1, 0, testDefinition())); err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	first := testDefinition()
	first.Name = "Primer cambio"
	if _, err := service.Publish(ctx, publication(1, 1, first)); err != nil {
		t.Fatalf("Publish() = %v", err)
	}

	second := testDefinition()
	second.Name = "Segundo cambio, editado desde la versión 1"
	if _, err := service.Publish(ctx, publication(1, 1, second)); !errors.Is(err, input.ErrOutdatedBase) {
		t.Errorf("Publish() = %v, want ErrOutdatedBase", err)
	}
	if v, err := service.Publish(ctx, publication(1, 1, first)); err != nil || v.Number != 2 {
		t.Errorf("Publish() of the latest content = version %d, %v, want version 2", v.Number, err)
	}
	if _, err := service.Publish(ctx, publication(3, 1, testDefinition())); !errors.Is(err, input.ErrOutdatedBase) {
		t.Errorf("Publish() from a base that does not exist = %v, want ErrOutdatedBase", err)
	}
}

func TestPublishAfterRenamingAnArea(t *testing.T) {
	ctx := context.Background()
	versions := memory.NewProcedureVersions()
	if _, err := newService(versions, applicant).Publish(ctx, publication(1, 0, testDefinition())); err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	renamed := applicant
	renamed.Name = "Estudiante"
	v, err := newService(versions, renamed).Publish(ctx, publication(1, 1, testDefinition()))
	if err != nil || v.Number != 2 {
		t.Errorf("Publish() = version %d, %v, want version 2", v.Number, err)
	}
	first, _ := versions.Get(ctx, 1, 1)
	if c, err := first.Content(); err != nil || c.Areas[0].Name != "Titular" {
		t.Errorf("version 1 areas = %v, %v, want the previous name", c.Areas, err)
	}
}

func TestPublishRejects(t *testing.T) {
	invalid := testDefinition()
	invalid.Steps[0].Area = "dean"
	tests := []struct {
		name    string
		service *ProcedureTypeService
		p       input.Publication
		want    error
	}{
		{"invalid definition", newService(memory.NewProcedureVersions(), applicant), publication(1, 0, invalid), procedure.ErrInvalidDefinition},
		{"invalid areas", newService(memory.NewProcedureVersions(), applicant, applicant), publication(1, 0, testDefinition()), procedure.ErrInvalidAreas},
		{"same proposer and approver", newService(memory.NewProcedureVersions(), applicant), input.Publication{
			ProcedureType: 1, Definition: testDefinition(), Authorship: procedure.Authorship{ProposedBy: "ana", ApprovedBy: "ana"},
		}, procedure.ErrInvalidAuthorship},
		{"area catalog fails", NewProcedureTypeService(memory.NewProcedureVersions(), failingCatalog{}, time.Now), publication(1, 0, testDefinition()), errStore},
		{"latest fails", newService(&fakeVersions{ProcedureVersions: memory.NewProcedureVersions(), latestErr: errStore}, applicant), publication(1, 0, testDefinition()), errStore},
		{"add fails", newService(&fakeVersions{ProcedureVersions: memory.NewProcedureVersions(), addErr: errStore}, applicant), publication(1, 0, testDefinition()), errStore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.service.Publish(context.Background(), tt.p); !errors.Is(err, tt.want) {
				t.Errorf("Publish() = %v, want %v", err, tt.want)
			}
		})
	}
}

// tamper stores version 1 of procedure type 1 with content that does not
// match its digest.
func tamper(t *testing.T, versions *memory.ProcedureVersions) {
	t.Helper()
	encoded := []byte(`{"format":1,"definition":{"name":"Alterado"}}`)
	v := procedure.Version{ProcedureType: 1, Number: 1, Digest: procedure.Digest([]byte("{}")), Encoded: encoded, Authorship: authors}
	if err := versions.Add(context.Background(), v); err != nil {
		t.Fatalf("Add() = %v", err)
	}
}

func TestAlteredVersion(t *testing.T) {
	ctx := context.Background()
	versions := memory.NewProcedureVersions()
	tamper(t, versions)
	service := newService(versions, applicant)
	if _, err := service.Publish(ctx, publication(1, 1, testDefinition())); !errors.Is(err, procedure.ErrAlteredContent) {
		t.Errorf("Publish() on an altered latest version = %v, want ErrAlteredContent", err)
	}
	if _, err := service.Version(ctx, 1, 1); !errors.Is(err, procedure.ErrAlteredContent) {
		t.Errorf("Version() of an altered version = %v, want ErrAlteredContent", err)
	}
	if _, err := service.Version(ctx, 1, 2); !errors.Is(err, output.ErrNoVersion) {
		t.Errorf("Version() of a missing version = %v, want ErrNoVersion", err)
	}
}

func TestVersionFromAnotherRow(t *testing.T) {
	ctx := context.Background()
	versions := &fakeVersions{ProcedureVersions: memory.NewProcedureVersions()}
	service := newService(versions, applicant)
	for _, procedureType := range []int64{1, 2} {
		if _, err := service.Publish(ctx, publication(procedureType, 0, testDefinition())); err != nil {
			t.Fatalf("Publish() = %v", err)
		}
	}
	versions.wrongRow = true
	if _, err := service.Version(ctx, 1, 1); err == nil || err.Error() != "procedure type 1 version 1: the store returned version 1 of procedure type 2" {
		t.Errorf("Version() = %v, want the wrong row rejected", err)
	}
}

// A stale Latest misses the last version, as if another publication had added
// it right after this one read the latest. Losing the race is rejected even
// with the same content: publishing again returns the stored version.
func TestConcurrentPublications(t *testing.T) {
	ctx := context.Background()
	versions := &fakeVersions{ProcedureVersions: memory.NewProcedureVersions()}
	service := newService(versions, applicant)
	if _, err := service.Publish(ctx, publication(1, 0, testDefinition())); err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	versions.stale = true
	if _, err := service.Publish(ctx, publication(1, 0, testDefinition())); !errors.Is(err, input.ErrOutdatedBase) {
		t.Errorf("Publish() after losing the race = %v, want ErrOutdatedBase", err)
	}
}
