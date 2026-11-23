package application_test

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
	"backend/src/auth/domain"
	"backend/src/auth/ports/input"
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

func TestOnlyAnAdministratorSearchesPeople(t *testing.T) {
	users := memory.NewUserRepository(applicant, operative)
	for i := range input.PeopleSearchLimit + 2 {
		users.Add(domain.User{ID: fmt.Sprintf("p%02d", i), Email: fmt.Sprintf("p%02d@fi.uba.ar", i), Name: fmt.Sprintf("Persona %02d", i)})
	}
	people := application.NewPeopleService(users, memory.NewAdministratorRepository(operative.ID))

	if found, err := people.Search(t.Context(), operative.ID, " TITU "); err != nil || !slices.Equal(found, []domain.User{applicant}) {
		t.Errorf("Search(TITU) = %+v, %v; want %+v", found, err, applicant)
	}
	found, err := people.Search(t.Context(), operative.ID, "persona")
	if err != nil || len(found) != input.PeopleSearchLimit || found[0].ID != "p00" {
		t.Errorf("Search(persona) = %+v, %v; want the first %d by name", found, err, input.PeopleSearchLimit)
	}
	for _, text := range []string{"o", " ", "nadie"} {
		if found, err := people.Search(t.Context(), operative.ID, text); err != nil || len(found) != 0 {
			t.Errorf("Search(%q) = %+v, %v; want nobody", text, found, err)
		}
	}
	if _, err := people.Search(t.Context(), applicant.ID, "titular"); !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("Search() by who is not an administrator = %v, want ErrAccessDenied", err)
	}
}
