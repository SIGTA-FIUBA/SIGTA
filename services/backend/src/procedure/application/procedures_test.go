package application

import (
	"context"
	"errors"
	"testing"

	authinput "backend/src/auth/ports/input"
	"backend/src/procedure/adapters/memory"
	"backend/src/procedure/domain/version"
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
	service := NewProcedureService(types, procedures, applicants{}, func() string { return "started" })

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
	if err := procedures.Add(ctx, replaced); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(ctx, "ana", replaced.ID); !errors.Is(err, version.ErrNotPinnedVersion) {
		t.Errorf("Open() of a procedure pinned to another digest = %v, want ErrNotPinnedVersion", err)
	}
	if _, err := service.List(ctx, "ana"); !errors.Is(err, version.ErrNotPinnedVersion) {
		t.Errorf("List() with a procedure pinned to another digest = %v, want ErrNotPinnedVersion", err)
	}
}
