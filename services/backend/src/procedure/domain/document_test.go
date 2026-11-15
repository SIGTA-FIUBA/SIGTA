package domain_test

import (
	"reflect"
	"testing"

	"backend/src/procedure/domain"
)

func TestCurrentKeepsTheLastOfEachRequirement(t *testing.T) {
	attached := []domain.Document{
		{ID: "1", Requirement: "transcript"},
		{ID: "2", Requirement: "letter"},
		{ID: "3", Requirement: "transcript"},
	}
	got := domain.Current(attached)
	want := []domain.Document{{ID: "2", Requirement: "letter"}, {ID: "3", Requirement: "transcript"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Current = %+v, want %+v", got, want)
	}
}
