package application_test

import (
	"errors"
	"testing"
	"time"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
)

func newAdministratorID() string { return "A-1" }

func TestGrantFirstAdministratorFindsTheUserByEmail(t *testing.T) {
	administrators := memory.NewAdministratorRepository()
	users := memory.NewUserRepository(applicant)

	user, err := application.GrantFirstAdministrator(t.Context(), users, administrators, " TITULAR@fi.uba.ar ", newAdministratorID, time.Now())
	if err != nil || user != applicant {
		t.Fatalf("GrantFirstAdministrator() = %+v, %v; want %+v", user, err, applicant)
	}
	if ok, _ := administrators.IsAdministrator(t.Context(), applicant.ID); !ok {
		t.Error("the user is not an administrator after GrantFirstAdministrator()")
	}
}

func TestGrantFirstAdministratorOnlyOnce(t *testing.T) {
	administrators := memory.NewAdministratorRepository()
	users := memory.NewUserRepository(applicant, operative)
	if _, err := application.GrantFirstAdministrator(t.Context(), users, administrators, applicant.Email, newAdministratorID, time.Now()); err != nil {
		t.Fatal(err)
	}

	_, err := application.GrantFirstAdministrator(t.Context(), users, administrators, operative.Email, newAdministratorID, time.Now())
	if !errors.Is(err, application.ErrAdministratorExists) {
		t.Errorf("GrantFirstAdministrator() with an administrator = %v, want ErrAdministratorExists", err)
	}
	if ok, _ := administrators.IsAdministrator(t.Context(), operative.ID); ok {
		t.Error("a second administrator was granted from the console")
	}
}

func TestGrantFirstAdministratorNeedsTheUserToHaveSignedIn(t *testing.T) {
	administrators := memory.NewAdministratorRepository()
	_, err := application.GrantFirstAdministrator(t.Context(), memory.NewUserRepository(applicant), administrators, operative.Email, newAdministratorID, time.Now())
	if !errors.Is(err, application.ErrUserNotSignedIn) {
		t.Errorf("GrantFirstAdministrator() of an email nobody signed in with = %v, want ErrUserNotSignedIn", err)
	}
	if ok, _ := administrators.IsAdministrator(t.Context(), operative.ID); ok {
		t.Error("GrantFirstAdministrator() granted it anyway")
	}
}

func TestAnAdministratorGrantsAndEndsAnother(t *testing.T) {
	// The memory repository tells who administers at the real time.
	now := time.Now().Add(-time.Hour)
	periods := memory.NewAdministratorRepository(operative.ID)
	service := application.NewAdministratorService(periods, memory.NewUserRepository(applicant, operative), newAdministratorID, func() time.Time { return now })

	a, err := service.Grant(t.Context(), operative.ID, applicant.ID)
	if err != nil || a.UserID != applicant.ID || a.GrantedBy != operative.ID || !a.From.Equal(now) {
		t.Fatalf("Grant() = %+v, %v", a, err)
	}
	if _, err := service.Grant(t.Context(), operative.ID, applicant.ID); !errors.Is(err, domain.ErrAlreadyAdministrator) {
		t.Errorf("Grant() twice = %v, want ErrAlreadyAdministrator", err)
	}
	if _, err := service.Grant(t.Context(), operative.ID, "nadie"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Grant() to who never signed in = %v, want ErrUserNotFound", err)
	}

	now = now.Add(time.Hour)
	if err := service.End(t.Context(), applicant.ID, "A-"+operative.ID); err != nil {
		t.Fatal(err)
	}
	if ok, _ := periods.IsAdministrator(t.Context(), operative.ID); ok {
		t.Error("the administration did not end")
	}
	if err := service.End(t.Context(), applicant.ID, a.ID); !errors.Is(err, domain.ErrLastAdministrator) {
		t.Errorf("End() of the last administrator = %v, want ErrLastAdministrator", err)
	}
	if _, err := service.List(t.Context(), operative.ID); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("List() by a former administrator = %v, want ErrAccessDenied", err)
	}
}
