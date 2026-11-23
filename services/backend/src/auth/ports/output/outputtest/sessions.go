package outputtest

import (
	"errors"
	"testing"
	"time"

	"backend/src/auth/domain"
	"backend/src/auth/ports/output"
)

// TestSessionRepository runs the contract of output.SessionRepository.
// newRepository returns an empty repository each time it is called.
func TestSessionRepository(t *testing.T, newRepository func(t *testing.T) output.SessionRepository) {
	t.Run("save, touch and delete", func(t *testing.T) { saveTouchAndDelete(t, newRepository(t)) })
	t.Run("missing sessions", func(t *testing.T) { missingSessions(t, newRepository(t)) })
}

// at is a time with microseconds, which PostgreSQL keeps.
var at = time.Date(2026, 11, 14, 10, 0, 0, 123456000, time.UTC)

func saveTouchAndDelete(t *testing.T, r output.SessionRepository) {
	ctx := t.Context()
	session := domain.Session{TokenHash: "hash-1", UserID: "U-1", LastSeenAt: at}
	other := domain.Session{TokenHash: "hash-2", UserID: "U-1", LastSeenAt: at}
	for _, s := range []domain.Session{session, other} {
		if err := r.Save(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	if got, err := r.Get(ctx, "hash-1"); err != nil || got != session {
		t.Errorf("Get() = %+v, %v; want %+v", got, err, session)
	}

	session.LastSeenAt = at.Add(time.Hour)
	if err := r.Touch(ctx, "hash-1", session.LastSeenAt); err != nil {
		t.Fatal(err)
	}
	if got, err := r.Get(ctx, "hash-1"); err != nil || got != session {
		t.Errorf("Get() after Touch() = %+v, %v; want %+v", got, err, session)
	}

	if err := r.Delete(ctx, "hash-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(ctx, "hash-1"); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("Get() after Delete() = %v, want ErrSessionNotFound", err)
	}
	if got, err := r.Get(ctx, "hash-2"); err != nil || got != other {
		t.Errorf("Get() of the other session = %+v, %v; want %+v", got, err, other)
	}
}

func missingSessions(t *testing.T, r output.SessionRepository) {
	ctx := t.Context()
	if _, err := r.Get(ctx, "hash-1"); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("Get() = %v, want ErrSessionNotFound", err)
	}
	if err := r.Touch(ctx, "hash-1", at); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Errorf("Touch() = %v, want ErrSessionNotFound", err)
	}
	if err := r.Delete(ctx, "hash-1"); err != nil {
		t.Errorf("Delete() = %v, want nil", err)
	}
}
