// Package outputtest holds the contract tests that every adapter of an output
// port must pass, whatever it stores the data in.
package outputtest

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"backend/src/procedure/domain/version"
	"backend/src/procedure/ports/output"
)

// publishers is how many publications of the same version race each other.
const publishers = 8

// TestProcedureVersions runs the contract of output.ProcedureVersions.
// newStore returns an empty store each time it is called.
func TestProcedureVersions(t *testing.T, newStore func(t *testing.T) output.ProcedureVersions) {
	t.Run("numbers without gaps", func(t *testing.T) { numbersWithoutGaps(t, newStore(t)) })
	t.Run("latest of each", func(t *testing.T) { latestOfEach(t, newStore(t)) })
	t.Run("content byte for byte", func(t *testing.T) { byteForByte(t, newStore(t)) })
	t.Run("no shared memory", func(t *testing.T) { noSharedMemory(t, newStore(t)) })
	t.Run("concurrent publications", func(t *testing.T) { concurrentPublications(t, newStore(t)) })
}

// Version builds a version of a made-up content whose digest matches it.
func Version(procedureType int64, number int, name string) version.ProcedureTypeVersion {
	encoded := []byte(`{"format":1,"definition":{"name":"` + name + `"}}`)
	return version.ProcedureTypeVersion{
		ProcedureTypeID: procedureType,
		Number:          number,
		Digest:          version.Digest(encoded),
		EncodedContent:  encoded,
		Authorship:      version.Authorship{ProposedBy: "ana", ApprovedBy: "beto"},
		PublishedAt:     time.Date(2026, 11, 1, 12, 0, number, 123456000, time.UTC),
	}
}

func numbersWithoutGaps(t *testing.T, s output.ProcedureVersions) {
	ctx := t.Context()
	if _, err := s.Latest(ctx, 1); !errors.Is(err, output.ErrNoVersion) {
		t.Errorf("Latest() of an empty store = %v, want ErrNoVersion", err)
	}
	for _, add := range []struct {
		number int
		want   string
	}{
		{0, "an error"},
		{1, "nil"},
		{1, "ErrVersionExists"},
		{3, "an error for skipping version 2"},
		{2, "nil"},
		{1, "ErrVersionExists"},
	} {
		err := s.Add(ctx, Version(1, add.number, fmt.Sprintf("Versión %d", add.number)))
		switch add.want {
		case "nil":
			if err != nil {
				t.Fatalf("Add(%d) = %v, want nil", add.number, err)
			}
		case "ErrVersionExists":
			if !errors.Is(err, output.ErrVersionExists) {
				t.Errorf("Add(%d) again = %v, want ErrVersionExists", add.number, err)
			}
		default:
			if err == nil || errors.Is(err, output.ErrVersionExists) {
				t.Errorf("Add(%d) = %v, want %s other than ErrVersionExists", add.number, err, add.want)
			}
		}
	}

	latest, err := s.Latest(ctx, 1)
	if err != nil || latest.Number != 2 || latest.Digest != Version(1, 2, "Versión 2").Digest {
		t.Errorf("Latest() = version %d, %v, want version 2", latest.Number, err)
	}
	got, err := s.Get(ctx, 1, 1)
	if err != nil || got.Digest != Version(1, 1, "Versión 1").Digest {
		t.Errorf("Get(1, 1) = version %d, %v, want version 1", got.Number, err)
	}
	all, err := s.LatestOfEach(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("LatestOfEach() = %+v, %v, want one version", all, err)
	}
	for name, v := range map[string]version.ProcedureTypeVersion{"Latest()": latest, "Get()": got, "LatestOfEach()": all[0]} {
		if v.PublishedAt.Location() != time.UTC {
			t.Errorf("%s returned PublishedAt in %s, want UTC", name, v.PublishedAt.Location())
		}
	}
	for _, number := range []int{0, 3} {
		if _, err := s.Get(ctx, 1, number); !errors.Is(err, output.ErrNoVersion) {
			t.Errorf("Get(1, %d) = %v, want ErrNoVersion", number, err)
		}
	}

	other := Version(2, 1, "Otro tipo")
	if err := s.Add(ctx, other); err != nil {
		t.Fatalf("Add() of another procedure type = %v", err)
	}
	if got, err := s.Get(ctx, 2, 1); err != nil || got.ProcedureTypeID != 2 || got.Digest != other.Digest {
		t.Errorf("Get(2, 1) = %+v, %v, want the version of procedure type 2", got, err)
	}
	if got, err := s.Latest(ctx, 1); err != nil || got.Number != 2 {
		t.Errorf("Latest() of procedure type 1 = version %d, %v, want version 2", got.Number, err)
	}
}

func latestOfEach(t *testing.T, s output.ProcedureVersions) {
	ctx := t.Context()
	if got, err := s.LatestOfEach(ctx); err != nil || len(got) != 0 {
		t.Fatalf("LatestOfEach() of an empty store = %v, %v, want nothing", got, err)
	}
	for _, add := range []struct {
		procedureType int64
		number        int
	}{{2, 1}, {1, 1}, {1, 2}} {
		if err := s.Add(ctx, Version(add.procedureType, add.number, "Trámite")); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.LatestOfEach(ctx)
	if err != nil || len(got) != 2 || got[0].ProcedureTypeID != 1 || got[0].Number != 2 || got[1].ProcedureTypeID != 2 || got[1].Number != 1 {
		t.Errorf("LatestOfEach() = %+v, %v, want version 2 of procedure type 1 and version 1 of procedure type 2", got, err)
	}
}

// byteForByte stores content that a decoder would change: keys out of order,
// spaces, escapes, characters outside ASCII and bytes that are not UTF-8.
func byteForByte(t *testing.T, s output.ProcedureVersions) {
	ctx := t.Context()
	contents := [][]byte{
		[]byte("{ \"definition\": {\"name\":\"Equivalencia de materias: año, acción, ñandú, Ünïcödé, 日本語, 🎓\", \"description\":\"caf\\u00e9\"},\n\t\"format\": 1 }"),
		{'{', 0xff, 0xfe, 0x00, '}'},
	}
	for i, content := range contents {
		want := version.ProcedureTypeVersion{
			ProcedureTypeID: 1,
			Number:          i + 1,
			Digest:          version.Digest(content),
			EncodedContent:  content,
			Authorship:      version.Authorship{ProposedBy: "Ñoño Pérez", ApprovedBy: "José Müller"},
			PublishedAt:     time.Date(2026, 11, 1, 12, 0, 0, 123456000, time.UTC),
		}
		if err := s.Add(ctx, want); err != nil {
			t.Fatalf("Add() = %v", err)
		}
		got, err := s.Get(ctx, 1, want.Number)
		if err != nil {
			t.Fatalf("Get() = %v", err)
		}
		if !bytes.Equal(got.EncodedContent, content) {
			t.Errorf("Get() returned content\n%q\nwant\n%q", got.EncodedContent, content)
		}
		if got.ProcedureTypeID != want.ProcedureTypeID || got.Number != want.Number || got.Digest != want.Digest ||
			got.Authorship != want.Authorship || !got.PublishedAt.Equal(want.PublishedAt) {
			t.Errorf("Get() = %+v, want %+v", got, want)
		}
		if err := got.VerifyDigest(); err != nil {
			t.Errorf("VerifyDigest() of the stored version = %v", err)
		}
	}
}

func noSharedMemory(t *testing.T, s output.ProcedureVersions) {
	ctx := t.Context()
	v := Version(1, 1, "Primera")
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
		t.Fatalf("Get() = %v", err)
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

// concurrentPublications races several publications of each number. Exactly
// one wins and the rest get ErrVersionExists, so the numbers have no gaps and
// the latest is the winner.
func concurrentPublications(t *testing.T, s output.ProcedureVersions) {
	ctx := t.Context()
	const rounds = 5
	for number := 1; number <= rounds; number++ {
		var (
			wg      sync.WaitGroup
			start   = make(chan struct{})
			results = make([]error, publishers)
		)
		for i := range publishers {
			wg.Go(func() {
				<-start
				results[i] = s.Add(ctx, Version(1, number, fmt.Sprintf("Versión %d de quien publica %d", number, i)))
			})
		}
		close(start)
		wg.Wait()

		winner := -1
		for i, err := range results {
			switch {
			case err == nil && winner < 0:
				winner = i
			case err == nil:
				t.Fatalf("version %d was added twice, by publishers %d and %d", number, winner, i)
			case !errors.Is(err, output.ErrVersionExists):
				t.Fatalf("publisher %d of version %d got %v, want nil or ErrVersionExists", i, number, err)
			}
		}
		if winner < 0 {
			t.Fatalf("no publisher added version %d", number)
		}
		latest, err := s.Latest(ctx, 1)
		want := Version(1, number, fmt.Sprintf("Versión %d de quien publica %d", number, winner))
		if err != nil || latest.Number != number || latest.Digest != want.Digest {
			t.Fatalf("Latest() = version %d with digest %s, %v, want the version %d of the winner, %s", latest.Number, latest.Digest, err, number, want.Digest)
		}
	}
	for number := 1; number <= rounds; number++ {
		if _, err := s.Get(ctx, 1, number); err != nil {
			t.Errorf("Get(1, %d) = %v, want the version", number, err)
		}
	}
}
