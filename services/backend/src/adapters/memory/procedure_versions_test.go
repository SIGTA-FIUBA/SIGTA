package memory

import (
	"context"
	"errors"
	"testing"

	"backend/src/domain/procedure"
	"backend/src/ports/output"
)

func version(number int, name string) procedure.Version {
	encoded := []byte(`{"format":1,"definition":{"name":"` + name + `"}}`)
	return procedure.Version{ProcedureType: 1, Number: number, Digest: procedure.Digest(encoded), Encoded: encoded}
}

func TestProcedureVersions(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureVersions()
	if _, err := s.Latest(ctx, 1); !errors.Is(err, output.ErrNoVersion) {
		t.Errorf("Latest() error = %v, want ErrNoVersion", err)
	}

	if err := s.Add(ctx, version(0, "Cero")); err == nil {
		t.Errorf("Add(0) = nil, want an error")
	}
	if err := s.Add(ctx, version(1, "Primera")); err != nil {
		t.Fatalf("Add(1) = %v", err)
	}
	if err := s.Add(ctx, version(1, "Repetida")); !errors.Is(err, output.ErrVersionExists) {
		t.Errorf("Add(1 again) error = %v, want ErrVersionExists", err)
	}
	if err := s.Add(ctx, version(3, "Salteada")); err == nil {
		t.Errorf("Add(3) = nil, want an error for skipping version 2")
	}
	if err := s.Add(ctx, version(2, "Segunda")); err != nil {
		t.Fatalf("Add(2) = %v", err)
	}

	got, err := s.Latest(ctx, 1)
	if err != nil || got.Number != 2 || got.Digest != version(2, "Segunda").Digest {
		t.Errorf("Latest() = version %d, %v, want version 2", got.Number, err)
	}
	if first, err := s.Get(ctx, 1, 1); err != nil || first.Digest != version(1, "Primera").Digest {
		t.Errorf("Get(1) = version %d, %v, want version 1", first.Number, err)
	}
	other := version(1, "Otro tipo")
	other.ProcedureType = 2
	if err := s.Add(ctx, other); err != nil {
		t.Fatalf("Add() of another procedure type = %v", err)
	}
	if got, err := s.Get(ctx, 2, 1); err != nil || got.ProcedureType != 2 || got.Digest != other.Digest {
		t.Errorf("Get(2, 1) = %+v, %v, want the version of procedure type 2", got, err)
	}
	for _, number := range []int{0, 3} {
		if _, err := s.Get(ctx, 1, number); !errors.Is(err, output.ErrNoVersion) {
			t.Errorf("Get(%d) error = %v, want ErrNoVersion", number, err)
		}
	}
}

func TestProcedureVersionsDoNotShareMemory(t *testing.T) {
	ctx := context.Background()
	s := NewProcedureVersions()
	v := version(1, "Primera")
	if err := s.Add(ctx, v); err != nil {
		t.Fatalf("Add() = %v", err)
	}
	v.Encoded[0] = 'x'
	latest, _ := s.Latest(ctx, 1)
	latest.Encoded[1] = 'x'
	got, _ := s.Get(ctx, 1, 1)
	got.Encoded[2] = 'x'
	if again, _ := s.Latest(ctx, 1); again.Verify() != nil {
		t.Errorf("changing an added or returned version changed the stored one")
	}
}
