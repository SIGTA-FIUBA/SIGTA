package application_test

import (
	"maps"
	"testing"

	"backend/src/auth/adapters/memory"
	"backend/src/auth/application"
)

func TestNamesLeavesOutWhoIsNotRegistered(t *testing.T) {
	people := application.NewPeopleService(memory.NewUserRepository(applicant, operative))

	names, err := people.Names(t.Context(), []string{applicant.ID, "nadie", applicant.ID, operative.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{applicant.ID: applicant.Name, operative.ID: operative.Name}
	if !maps.Equal(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}
