package application_test

import (
	"errors"
	"testing"
	"time"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
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
