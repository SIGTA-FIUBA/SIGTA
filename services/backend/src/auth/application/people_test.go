package application_test

import (
	"errors"
	"maps"
	"testing"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
)

func TestNamesLeavesOutWhoIsNotRegistered(t *testing.T) {
	people := application.NewPeopleService(memory.NewUserRepository(applicant, operative), memory.NewAdministratorRepository())

	names, err := people.Names(t.Context(), []string{applicant.ID, "nadie", applicant.ID, operative.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{applicant.ID: applicant.Name, operative.ID: operative.Name}
	if !maps.Equal(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}

func TestOnlyAnAdministratorFindsAPersonByEmail(t *testing.T) {
	people := application.NewPeopleService(memory.NewUserRepository(applicant, operative), memory.NewAdministratorRepository(operative.ID))

	if user, err := people.Find(t.Context(), operative.ID, " Titular@FI.uba.ar"); err != nil || user != applicant {
		t.Errorf("Find() = %+v, %v; want %+v", user, err, applicant)
	}
	if _, err := people.Find(t.Context(), operative.ID, "nadie@fi.uba.ar"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Find() of who never signed in = %v, want ErrUserNotFound", err)
	}
	if _, err := people.Find(t.Context(), applicant.ID, operative.Email); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Find() by who is not an administrator = %v, want ErrAccessDenied", err)
	}
}
