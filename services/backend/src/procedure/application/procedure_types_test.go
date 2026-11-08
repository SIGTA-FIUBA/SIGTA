package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	authmem "backend/src/auth/adapters/memory"
	authapp "backend/src/auth/application"
	authdomain "backend/src/auth/domain"
	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/domain"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

var (
	office      = authdomain.Area{ID: "office", Name: "Secretaría"}
	publishedAt = time.Date(2026, 11, 1, 12, 0, 0, 123456789, time.UTC)
	authors     = domain.Authorship{ProposedBy: "ana", ApprovedBy: "beto"}
	errStore    = errors.New("store unavailable")
)

func testDefinition() domain.Definition {
	return domain.Definition{
		Name:        "Trámite de prueba",
		Description: "Un trámite para las pruebas",
		Fields:      []domain.Field{{ID: "national_id", Name: "DNI", Type: domain.FieldText}},
		Steps: []domain.Step{
			{ID: "request", Name: "Solicitud", Role: domain.RoleApplicant, Action: domain.ActionEnterData, Fields: []string{"national_id"}},
		},
	}
}

func publication(procedureType int64, base int, d domain.Definition) input.Publication {
	return input.Publication{ProcedureType: procedureType, Base: base, Definition: d, Authorship: authors}
}

// newService reads the staff areas from auth, as the server does (ADR-018).
func newService(versions output.ProcedureVersions, areas ...authdomain.Area) *ProcedureTypeService {
	return newServiceOn(versions, authmem.NewAreaRepository(areas...))
}

func newServiceOn(versions output.ProcedureVersions, areas *authmem.AreaRepository) *ProcedureTypeService {
	auth := authapp.NewAreaService(areas, authmem.NewAdministrators(), authmem.NewMembershipRepository(), authmem.NewDeniedAttemptLog(), time.Now)
	return NewProcedureTypeService(versions, auth, func() time.Time { return publishedAt })
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

func (f *fakeVersions) Get(ctx context.Context, procedureType int64, number int) (domain.Version, error) {
	if f.wrongRow {
		return f.ProcedureVersions.Get(ctx, procedureType+1, number)
	}
	return f.ProcedureVersions.Get(ctx, procedureType, number)
}

func (f *fakeVersions) Latest(ctx context.Context, procedureType int64) (domain.Version, error) {
	if f.latestErr != nil {
		return domain.Version{}, f.latestErr
	}
	if f.stale {
		f.stale = false
		return domain.Version{}, output.ErrNoVersion
	}
	return f.ProcedureVersions.Latest(ctx, procedureType)
}

func (f *fakeVersions) Add(ctx context.Context, v domain.Version) error {
	if f.addErr != nil {
		return f.addErr
	}
	return f.ProcedureVersions.Add(ctx, v)
}

type failingAreas struct{ authinput.Areas }

func (failingAreas) List(context.Context) ([]authdomain.Area, error) { return nil, errStore }

func TestPublish(t *testing.T) {
	ctx := context.Background()
	service := newService(memory.NewProcedureVersions())
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
	again.Authorship = domain.Authorship{ProposedBy: "carla", ApprovedBy: "dario"}
	stored, err := service.Version(ctx, 1, 1)
	if err != nil {
		t.Fatalf("Version() = %v", err)
	}
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
	service := newService(memory.NewProcedureVersions())
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
	_, err := service.Publish(ctx, publication(1, 1, second))
	if outdated, ok := errors.AsType[*input.OutdatedBaseError](err); !ok || outdated.Latest != 2 || !errors.Is(err, input.ErrOutdatedBase) {
		t.Errorf("Publish() = %v, want ErrOutdatedBase with latest version 2", err)
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
	areas := authmem.NewAreaRepository(office)
	reviewed := testDefinition()
	reviewed.Steps[0].Next = "review"
	reviewed.Steps = append(reviewed.Steps, domain.Step{ID: "review", Name: "Revisión", Role: domain.RoleStaff, Area: "office", Profile: domain.ProfileStaffOperative, Action: domain.ActionConfirmExternalFact})
	if _, err := newServiceOn(versions, areas).Publish(ctx, publication(1, 0, reviewed)); err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	renamed := office
	renamed.Name = "Dirección de Alumnos"
	if err := areas.Save(ctx, renamed); err != nil {
		t.Fatal(err)
	}
	v, err := newServiceOn(versions, areas).Publish(ctx, publication(1, 1, reviewed))
	if err != nil || v.Number != 2 {
		t.Errorf("Publish() = version %d, %v, want version 2", v.Number, err)
	}
	first, err := versions.Get(ctx, 1, 1)
	if err != nil {
		t.Fatalf("Get() = %v", err)
	}
	c, err := first.Content()
	if err != nil {
		t.Fatalf("Content() = %v", err)
	}
	if want := []domain.Area{{ID: "office", Name: "Secretaría"}}; !reflect.DeepEqual(c.Areas, want) {
		t.Errorf("version 1 areas = %v, want the previous name %v", c.Areas, want)
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
		{name: "invalid definition", service: newService(memory.NewProcedureVersions()), p: publication(1, 0, invalid), want: domain.ErrInvalidDefinition},
		{name: "area stored with an invalid id", service: newService(memory.NewProcedureVersions(), authdomain.Area{ID: "Office", Name: "Secretaría"}), p: publication(1, 0, testDefinition()), want: domain.ErrInvalidAreas},
		{name: "same proposer and approver", service: newService(memory.NewProcedureVersions()), p: input.Publication{
			ProcedureType: 1, Definition: testDefinition(), Authorship: domain.Authorship{ProposedBy: "ana", ApprovedBy: "ana"},
		}, want: domain.ErrInvalidAuthorship},
		{name: "area catalog fails", service: NewProcedureTypeService(memory.NewProcedureVersions(), failingAreas{}, time.Now), p: publication(1, 0, testDefinition()), want: errStore},
		{name: "latest fails", service: newService(&fakeVersions{ProcedureVersions: memory.NewProcedureVersions(), latestErr: errStore}), p: publication(1, 0, testDefinition()), want: errStore},
		{name: "add fails", service: newService(&fakeVersions{ProcedureVersions: memory.NewProcedureVersions(), addErr: errStore}), p: publication(1, 0, testDefinition()), want: errStore},
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
	v := domain.Version{ProcedureType: 1, Number: 1, Digest: domain.Digest([]byte("{}")), Encoded: encoded, Authorship: authors}
	if err := versions.Add(context.Background(), v); err != nil {
		t.Fatalf("Add() = %v", err)
	}
}

func TestAlteredVersion(t *testing.T) {
	ctx := context.Background()
	versions := memory.NewProcedureVersions()
	tamper(t, versions)
	service := newService(versions)
	if _, err := service.Publish(ctx, publication(1, 1, testDefinition())); !errors.Is(err, domain.ErrAlteredContent) {
		t.Errorf("Publish() on an altered latest version = %v, want ErrAlteredContent", err)
	}
	if _, err := service.Version(ctx, 1, 1); !errors.Is(err, domain.ErrAlteredContent) {
		t.Errorf("Version() of an altered version = %v, want ErrAlteredContent", err)
	}
	if _, err := service.Version(ctx, 1, 2); !errors.Is(err, output.ErrNoVersion) {
		t.Errorf("Version() of a missing version = %v, want ErrNoVersion", err)
	}
}

func TestVersionFromAnotherRow(t *testing.T) {
	ctx := context.Background()
	versions := &fakeVersions{ProcedureVersions: memory.NewProcedureVersions()}
	service := newService(versions)
	for _, procedureType := range []int64{1, 2} {
		if _, err := service.Publish(ctx, publication(procedureType, 0, testDefinition())); err != nil {
			t.Fatalf("Publish() = %v", err)
		}
	}
	versions.wrongRow = true
	if _, err := service.Version(ctx, 1, 1); !errors.Is(err, output.ErrWrongVersion) {
		t.Errorf("Version() = %v, want ErrWrongVersion", err)
	}
}

// A stale Latest misses the last version, as if another publication had added
// it right after this one read the latest. Losing the race is rejected even
// with the same content: publishing again returns the stored version.
func TestConcurrentPublications(t *testing.T) {
	ctx := context.Background()
	versions := &fakeVersions{ProcedureVersions: memory.NewProcedureVersions()}
	service := newService(versions)
	if _, err := service.Publish(ctx, publication(1, 0, testDefinition())); err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	versions.stale = true
	_, err := service.Publish(ctx, publication(1, 0, testDefinition()))
	if outdated, ok := errors.AsType[*input.OutdatedBaseError](err); !ok || outdated.Latest != 1 {
		t.Errorf("Publish() after losing the race = %v, want ErrOutdatedBase with latest version 1", err)
	}
}

func TestCatalog(t *testing.T) {
	ctx := context.Background()
	service := newService(memory.NewProcedureVersions())
	changed := testDefinition()
	changed.Name = "Trámite cambiado"
	for _, p := range []input.Publication{
		publication(1, 0, testDefinition()),
		publication(1, 1, changed),
		publication(2, 0, testDefinition()),
	} {
		if _, err := service.Publish(ctx, p); err != nil {
			t.Fatalf("Publish() = %v", err)
		}
	}

	entries, err := service.Catalog(ctx)
	if err != nil {
		t.Fatalf("Catalog() = %v", err)
	}
	want := []struct {
		procedureType int64
		version       int
		name          string
	}{{1, 2, "Trámite cambiado"}, {2, 1, "Trámite de prueba"}}
	if len(entries) != len(want) {
		t.Fatalf("Catalog() = %+v, want %+v", entries, want)
	}
	for index, w := range want {
		entry := entries[index]
		if entry.ProcedureType != w.procedureType || entry.Version != w.version || entry.Name != w.name {
			t.Errorf("entries[%d] = %+v, want %+v", index, entry, w)
		}
		if len(entry.Fields) != 1 || !entry.Fields[0].Required {
			t.Errorf("entries[%d].Fields = %+v, want national_id as mandatory", index, entry.Fields)
		}
	}
}

func TestCatalogWithUnreadableContent(t *testing.T) {
	versions := memory.NewProcedureVersions()
	encoded := []byte(`{"format":99}`)
	v := domain.Version{ProcedureType: 1, Number: 1, Digest: domain.Digest(encoded), Encoded: encoded, Authorship: authors}
	if err := versions.Add(context.Background(), v); err != nil {
		t.Fatalf("Add() = %v", err)
	}
	if entries, err := newService(versions).Catalog(context.Background()); err == nil {
		t.Errorf("Catalog() = %+v, want an error", entries)
	}
}
