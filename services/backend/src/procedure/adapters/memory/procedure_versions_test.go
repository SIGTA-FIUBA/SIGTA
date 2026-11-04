package memory

import (
	"context"
	"errors"
	"testing"

	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/output"
)

func testVersion(number int, name string) version.ProcedureTypeVersion {
	encoded := []byte(`{"format":1,"definition":{"name":"` + name + `"}}`)
	return version.ProcedureTypeVersion{ProcedureTypeID: 1, Number: number, Digest: version.Digest(encoded), EncodedContent: encoded}
}

func TestProcedureVersions(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureVersions()
	if _, err := s.Latest(ctx, 1); !errors.Is(err, output.ErrNoVersion) {
		t.Errorf("Latest() error = %v, want ErrNoVersion", err)
	}

	if err := s.Add(ctx, testVersion(0, "Cero")); err == nil {
		t.Errorf("Add(0) = nil, want an error")
	}
	if err := s.Add(ctx, testVersion(1, "Primera")); err != nil {
		t.Fatalf("Add(1) = %v", err)
	}
	if err := s.Add(ctx, testVersion(1, "Repetida")); !errors.Is(err, output.ErrVersionExists) {
		t.Errorf("Add(1 again) error = %v, want ErrVersionExists", err)
	}
	if err := s.Add(ctx, testVersion(3, "Salteada")); err == nil {
		t.Errorf("Add(3) = nil, want an error for skipping version 2")
	}
	if err := s.Add(ctx, testVersion(2, "Segunda")); err != nil {
		t.Fatalf("Add(2) = %v", err)
	}

	got, err := s.Latest(ctx, 1)
	if err != nil || got.Number != 2 || got.Digest != testVersion(2, "Segunda").Digest {
		t.Errorf("Latest() = version %d, %v, want version 2", got.Number, err)
	}
	if first, err := s.Get(ctx, 1, 1); err != nil || first.Digest != testVersion(1, "Primera").Digest {
		t.Errorf("Get(1) = version %d, %v, want version 1", first.Number, err)
	}
	other := testVersion(1, "Otro tipo")
	other.ProcedureTypeID = 2
	if err := s.Add(ctx, other); err != nil {
		t.Fatalf("Add() of another procedure type = %v", err)
	}
	if got, err := s.Get(ctx, 2, 1); err != nil || got.ProcedureTypeID != 2 || got.Digest != other.Digest {
		t.Errorf("Get(2, 1) = %+v, %v, want the version of procedure type 2", got, err)
	}
	for _, number := range []int{0, 3} {
		if _, err := s.Get(ctx, 1, number); !errors.Is(err, output.ErrNoVersion) {
			t.Errorf("Get(%d) error = %v, want ErrNoVersion", number, err)
		}
	}
}

func TestLatestOfEach(t *testing.T) {
	ctx := t.Context()
	s := NewProcedureVersions()
	if got, err := s.LatestOfEach(ctx); err != nil || len(got) != 0 {
		t.Fatalf("LatestOfEach() = %v, %v, want nothing", got, err)
	}
	for _, add := range []struct {
		procedureType int64
		number        int
	}{{2, 1}, {1, 1}, {1, 2}} {
		added := testVersion(add.number, "Trámite")
		added.ProcedureTypeID = add.procedureType
		if err := s.Add(ctx, added); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.LatestOfEach(ctx)
	if err != nil || len(got) != 2 || got[0].ProcedureTypeID != 1 || got[0].Number != 2 || got[1].ProcedureTypeID != 2 || got[1].Number != 1 {
		t.Errorf("LatestOfEach() = %+v, %v", got, err)
	}
}

func TestProcedureVersionsDoNotShareMemory(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureVersions()
	v := testVersion(1, "Primera")
	if err := s.Add(ctx, v); err != nil {
		t.Fatalf("Add() = %v", err)
	}
	v.EncodedContent[0] = 'x'
	latest, err := s.Latest(ctx, 1)
	if err != nil {
		t.Fatalf("Latest() = %v", err)
	}
	latest.EncodedContent[1] = 'x'
	got, err := s.Get(ctx, 1, 1)
	if err != nil {
		t.Fatalf("Get(1) = %v", err)
	}
	got.EncodedContent[2] = 'x'
	again, err := s.Latest(ctx, 1)
	if err != nil {
		t.Fatalf("Latest() = %v", err)
	}
	if err := again.VerifyDigest(); err != nil {
		t.Errorf("changing an added or returned version changed the stored one: %v", err)
	}
}
