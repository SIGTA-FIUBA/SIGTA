// Package outputtest has the contract tests of the output ports of the auth
// module: each adapter runs them with an empty repository of its own.
package outputtest

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

// TestUserRepository runs the contract of output.UserRepository.
// newRepository returns an empty repository each time it is called.
func TestUserRepository(t *testing.T, newRepository func(t *testing.T) output.UserRepository) {
	t.Run("register and find", func(t *testing.T) { registerAndFind(t, newRepository(t)) })
	t.Run("one user per email", func(t *testing.T) { oneUserPerEmail(t, newRepository(t)) })
	t.Run("concurrent first logins", func(t *testing.T) { concurrentFirstLogins(t, newRepository(t)) })
	t.Run("search", func(t *testing.T) { search(t, newRepository(t)) })
}

func registerAndFind(t *testing.T, r output.UserRepository) {
	ctx := t.Context()
	if _, err := r.Get(ctx, "U-1"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Get() of an empty repository = %v, want ErrUserNotFound", err)
	}
	if _, err := r.GetByEmail(ctx, "ana@fi.uba.ar"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("GetByEmail() of an empty repository = %v, want ErrUserNotFound", err)
	}
	want := domain.User{ID: "U-1", Email: "ana@fi.uba.ar", Name: "Ana Pérez"}
	if got, err := r.Register(ctx, domain.User{ID: "U-1", Email: " Ana@FI.uba.ar ", Name: "Ana Pérez"}); err != nil || got != want {
		t.Fatalf("Register() = %+v, %v; want %+v", got, err, want)
	}
	if got, err := r.Get(ctx, "U-1"); err != nil || got != want {
		t.Errorf("Get() = %+v, %v; want %+v", got, err, want)
	}
	if got, err := r.GetByEmail(ctx, "ana@fi.uba.ar"); err != nil || got != want {
		t.Errorf("GetByEmail() = %+v, %v; want %+v", got, err, want)
	}
}

func oneUserPerEmail(t *testing.T, r output.UserRepository) {
	ctx := t.Context()
	first := domain.User{ID: "U-1", Email: "ana@fi.uba.ar", Name: "Ana Pérez"}
	if _, err := r.Register(ctx, first); err != nil {
		t.Fatal(err)
	}
	if got, err := r.Register(ctx, domain.User{ID: "U-2", Email: "ANA@fi.uba.ar", Name: "Ana"}); err != nil || got != first {
		t.Errorf("Register() of a registered email = %+v, %v; want %+v", got, err, first)
	}
	if _, err := r.Get(ctx, "U-2"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Get(U-2) = %v, want ErrUserNotFound", err)
	}
	if _, err := r.Register(ctx, domain.User{ID: "U-1", Email: "beto@fi.uba.ar", Name: "Beto"}); err == nil {
		t.Error("Register() of an id taken by another email = nil, want an error")
	}
	if got, err := r.Get(ctx, "U-1"); err != nil || got != first {
		t.Errorf("Get(U-1) = %+v, %v; want %+v", got, err, first)
	}
}

func concurrentFirstLogins(t *testing.T, r output.UserRepository) {
	const logins = 10
	var wg sync.WaitGroup
	got := make([]domain.User, logins)
	errs := make([]error, logins)
	for i := range logins {
		wg.Go(func() {
			got[i], errs[i] = r.Register(t.Context(), domain.User{ID: fmt.Sprintf("U-%d", i), Email: "ana@fi.uba.ar", Name: "Ana"})
		})
	}
	wg.Wait()
	for i := range logins {
		if errs[i] != nil || got[i] != got[0] {
			t.Errorf("Register() number %d = %+v, %v; want %+v like the others", i, got[i], errs[i], got[0])
		}
	}
}

func search(t *testing.T, r output.UserRepository) {
	ctx := t.Context()
	ana := domain.User{ID: "U-1", Email: "ana@fi.uba.ar", Name: "Ana Pérez"}
	alvaro := domain.User{ID: "U-2", Email: "alvaro@fi.uba.ar", Name: "ÁLVARO GÓMEZ"}
	luz := domain.User{ID: "U-3", Email: "luz_rey@fi.uba.ar", Name: "Luz Rey"}
	lucia := domain.User{ID: "U-4", Email: "luzarey@fi.uba.ar", Name: "Lucía Rey"}
	joao := domain.User{ID: "U-5", Email: "joao@example.org", Name: "João Conceição"}
	ozge := domain.User{ID: "U-6", Email: "ozge@example.org", Name: "ÖZGE ÇELIK"}
	slash := domain.User{ID: "U-7", Email: "slash@example.org", Name: `Back\Slash`}
	for _, u := range []domain.User{luz, ana, lucia, alvaro, joao, ozge, slash} {
		if _, err := r.Register(ctx, u); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		text  string
		limit int
		want  []domain.User
	}{
		{"perez", 10, []domain.User{ana}},
		{"alvaro gomez", 10, []domain.User{alvaro}},
		{"fi.uba", 10, []domain.User{alvaro, ana, lucia, luz}},
		{"fi.uba", 2, []domain.User{alvaro, ana}},
		{"conceicao", 10, []domain.User{joao}},
		{"ozge cel", 10, []domain.User{ozge}},
		{"example", 10, []domain.User{slash, joao, ozge}},
		{"luz_", 10, []domain.User{luz}},
		{`k\s`, 10, []domain.User{slash}},
		{"%", 10, nil},
		{"nadie", 10, nil},
	} {
		got, err := r.Search(ctx, tc.text, tc.limit)
		if err != nil || len(got) != len(tc.want) || (len(got) > 0 && !slices.Equal(got, tc.want)) {
			t.Errorf("Search(%q, %d) = %+v, %v; want %+v", tc.text, tc.limit, got, err, tc.want)
		}
	}
}
