package application

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	procdomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/input"
)

// applicants answers that nobody is staff.
type applicants struct{}

func (applicants) IsStaff(context.Context, string) (bool, error) { return false, nil }

func (applicants) Holds(context.Context, string, string, int64, []string) (bool, error) {
	return false, nil
}

func (applicants) Reach(context.Context, string) ([]authinput.AreaReach, error) { return nil, nil }

// A procedure follows the version it started with, number and digest: a
// version lost and published again with the same number is not that one.
func TestProcedureFollowsItsPinnedVersion(t *testing.T) {
	ctx := t.Context()
	types := newService(memory.NewProcedureVersions(), office)
	published, err := types.Publish(ctx, publication(1, 0, testDefinition()))
	if err != nil {
		t.Fatal(err)
	}
	procedures := memory.NewProcedureRepository()
	service := NewProcedureService(types, procedures, memory.NewDocumentRepository(), applicants{}, func() string { return "started" }, time.Now)

	started, err := service.Start(ctx, "ana", 1)
	if err != nil {
		t.Fatal(err)
	}
	if started.Pin() != published.Pin() {
		t.Fatalf("Start() pins %+v, want %+v", started.Pin(), published.Pin())
	}
	if _, err := service.Open(ctx, "ana", started.ID); err != nil {
		t.Fatalf("Open() = %v, want nil", err)
	}

	replaced := started
	replaced.ID, replaced.VersionDigest = "replaced", "sha256:00"
	if err := procedures.Add(ctx, replaced, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(ctx, "ana", replaced.ID); !errors.Is(err, version.ErrNotPinnedVersion) {
		t.Errorf("Open() of a procedure pinned to another digest = %v, want ErrNotPinnedVersion", err)
	}
	if _, err := service.List(ctx, "ana"); !errors.Is(err, version.ErrNotPinnedVersion) {
		t.Errorf("List() with a procedure pinned to another digest = %v, want ErrNotPinnedVersion", err)
	}
}

var confirmedAt = time.Date(2026, 11, 16, 10, 30, 0, 0, time.UTC)

type fixedVersion struct {
	input.ProcedureTypes
	encoded []byte
}

func (f fixedVersion) Pinned(context.Context, version.Pin) (version.ProcedureTypeVersion, error) {
	return version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: 1, Digest: version.Digest(f.encoded), EncodedContent: f.encoded}, nil
}

type staffHolding map[string]map[string][]string

func (s staffHolding) Holds(_ context.Context, userID, area string, _ int64, permissions []string) (bool, error) {
	held := s[userID][area]
	for _, p := range permissions {
		if !slices.Contains(held, p) {
			return false, nil
		}
	}
	return held != nil, nil
}

func (s staffHolding) Reach(_ context.Context, userID string) ([]authinput.AreaReach, error) {
	var reach []authinput.AreaReach
	for area := range s[userID] {
		reach = append(reach, authinput.AreaReach{Area: area})
	}
	return reach, nil
}

type failingAdvance struct {
	*memory.ProcedureRepository
}

func (failingAdvance) Advance(context.Context, procdomain.Procedure, procdomain.Event) error {
	return errStore
}

func confirmationFixture(t *testing.T) (*ProcedureService, *memory.ProcedureRepository) {
	t.Helper()
	d := definition.ProcedureTypeDefinition{
		Name: "Equivalencia",
		Steps: []definition.Step{
			{ID: "request", Role: definition.RoleApplicant, Action: definition.ActionEnterData, Next: "file"},
			{ID: "file", Role: definition.RoleStaff, Area: "affairs", Permissions: []string{"verify"}, Action: definition.ActionConfirmExternalFact, Next: "evaluate"},
			{ID: "evaluate", Role: definition.RoleStaff, Area: "commission", Permissions: []string{"evaluate"}, Action: definition.ActionConfirmExternalFact},
		},
	}
	encoded, err := version.NewContent(d, nil).Encode()
	if err != nil {
		t.Fatal(err)
	}
	staff := staffHolding{
		"operative@fi.uba.ar":  {"affairs": {"verify"}},
		"commission@fi.uba.ar": {"commission": {"evaluate"}},
	}
	procedures := memory.NewProcedureRepository()
	service := NewProcedureService(fixedVersion{encoded: encoded}, procedures, memory.NewDocumentRepository(), staff, func() string { return "P-1" }, func() time.Time { return confirmedAt })
	if err := procedures.Add(context.Background(), procdomain.Procedure{ID: "P-1", ProcedureType: 1, Version: 1, Applicant: "applicant@fi.uba.ar", OpenStep: "file"}, nil); err != nil {
		t.Fatal(err)
	}
	return service, procedures
}

func fileConfirmation() input.StepConfirmation {
	return input.StepConfirmation{Step: "file", Observation: "case file created", ExternalFileNumber: "EX-1"}
}

func TestConfirmAdvancesAndRecordsWhoAndWhen(t *testing.T) {
	ctx := context.Background()
	service, procedures := confirmationFixture(t)

	event, err := service.Confirm(ctx, "operative@fi.uba.ar", "P-1", fileConfirmation())
	if err != nil {
		t.Fatal(err)
	}
	if event.By != "operative@fi.uba.ar" || !event.At.Equal(confirmedAt) || event.Observation != "case file created" {
		t.Errorf("event = %+v, want it by operative at %v with the observation", event, confirmedAt)
	}
	p, err := procedures.Get(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if p.OpenStep != "evaluate" || p.Referral == nil || p.Referral.To != "commission" {
		t.Errorf("procedure = %+v, want it on evaluate, referred to commission", p)
	}
	events, err := procedures.Events(ctx, "P-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Referral == nil || events[0].Referral.ExternalFileNumber != "EX-1" {
		t.Errorf("events = %+v, want one referral with its external file number", events)
	}
}

func TestConfirmRejectsAndChangesNothing(t *testing.T) {
	notOpen := fileConfirmation()
	notOpen.Step = "evaluate"
	noObservation := fileConfirmation()
	noObservation.Observation = ""
	applicantStep := input.StepConfirmation{Step: "request", Observation: "sent"}

	tests := []struct {
		name         string
		user         string
		confirmation input.StepConfirmation
		openStep     string
		want         error
	}{
		{"staff without the permissions of the step", "commission@fi.uba.ar", fileConfirmation(), "file", procdomain.ErrAccessDenied},
		{"the applicant on a staff step", "applicant@fi.uba.ar", fileConfirmation(), "file", procdomain.ErrAccessDenied},
		{"someone who cannot see the procedure", "other@fi.uba.ar", fileConfirmation(), "file", procdomain.ErrProcedureNotFound},
		{"a step that is not open, even without its permissions", "operative@fi.uba.ar", notOpen, "file", engine.ErrStepNotOpen},
		{"no observation", "operative@fi.uba.ar", noObservation, "file", engine.ErrNoObservation},
		{"a step of the applicant", "applicant@fi.uba.ar", applicantStep, "request", engine.ErrApplicantStep},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			service, procedures := confirmationFixture(t)
			if tt.openStep != "file" {
				if err := procedures.Add(ctx, procdomain.Procedure{ID: "P-1", ProcedureType: 1, Version: 1, Applicant: "applicant@fi.uba.ar", OpenStep: tt.openStep}, nil); err != nil {
					t.Fatal(err)
				}
			}

			if _, err := service.Confirm(ctx, tt.user, "P-1", tt.confirmation); !errors.Is(err, tt.want) {
				t.Fatalf("Confirm = %v, want %v", err, tt.want)
			}
			p, _ := procedures.Get(ctx, "P-1")
			events, _ := procedures.Events(ctx, "P-1")
			if p.OpenStep != tt.openStep || len(events) != 0 {
				t.Errorf("open step = %q with %d events, want %q with none", p.OpenStep, len(events), tt.openStep)
			}
		})
	}
}

func TestConfirmTheSameStepTwice(t *testing.T) {
	ctx := context.Background()
	service, _ := confirmationFixture(t)
	if _, err := service.Confirm(ctx, "operative@fi.uba.ar", "P-1", fileConfirmation()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Confirm(ctx, "operative@fi.uba.ar", "P-1", fileConfirmation()); !errors.Is(err, engine.ErrStepNotOpen) {
		t.Errorf("second Confirm = %v, want ErrStepNotOpen", err)
	}
}

func TestConfirmWhenStoringFailsLeavesTheProcedureAsItWas(t *testing.T) {
	ctx := context.Background()
	service, procedures := confirmationFixture(t)
	service.procedures = failingAdvance{procedures}

	if _, err := service.Confirm(ctx, "operative@fi.uba.ar", "P-1", fileConfirmation()); !errors.Is(err, errStore) {
		t.Fatalf("Confirm = %v, want the store error", err)
	}
	p, _ := procedures.Get(ctx, "P-1")
	events, _ := procedures.Events(ctx, "P-1")
	if p.OpenStep != "file" || len(events) != 0 {
		t.Errorf("open step = %q with %d events, want file with none", p.OpenStep, len(events))
	}
}
