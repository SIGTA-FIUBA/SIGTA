package rest_test

import (
	"net/http"
	"strings"
	"testing"

	"backend/src/api"
)

func emailOf(p api.GrantedPerson) string {
	if p.Email == nil {
		return ""
	}
	return string(*p.Email)
}

func TestAnAdministratorSearchesPeople(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := get(t, admin, srv.URL+"/people?q=TITU")
	var people []api.RegisteredPerson
	decode(t, res, err, http.StatusOK, &people)
	if len(people) != 1 || people[0].Id != "u1" || people[0].Name != "Titular" || people[0].Email != "titular@fi.uba.ar" {
		t.Fatalf("people = %+v", people)
	}
	res, err = get(t, admin, srv.URL+"/people?q=nadie")
	decode(t, res, err, http.StatusOK, &people)
	if len(people) != 0 {
		t.Errorf("people = %+v, want nobody", people)
	}
}

func TestAnAdministratorGrantsAndEndsTheAdministration(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := remove(t, admin, srv.URL+"/administrators/A-u3")
	if msg := expectError(t, res, err, http.StatusConflict); !strings.Contains(msg, "única") {
		t.Errorf("message = %q, want it to say why", msg)
	}

	res, err = post(t, admin, srv.URL+"/administrators", strings.NewReader(`{"user_id": "u1"}`))
	var granted api.Administrator
	decode(t, res, err, http.StatusCreated, &granted)
	if granted.Person.Id != "u1" || granted.GrantedBy == nil || granted.GrantedBy.Id != "u3" {
		t.Fatalf("administrator = %+v", granted)
	}
	res, err = post(t, admin, srv.URL+"/administrators", strings.NewReader(`{"user_id": "u1"}`))
	expectError(t, res, err, http.StatusConflict)
	res, err = post(t, admin, srv.URL+"/administrators", strings.NewReader(`{"user_id": "nadie"}`))
	expectError(t, res, err, http.StatusBadRequest)

	res, err = get(t, admin, srv.URL+"/administrators")
	var administrators []api.Administrator
	decode(t, res, err, http.StatusOK, &administrators)
	if len(administrators) != 2 || administrators[0].Id != "A-u3" || administrators[0].GrantedBy != nil || administrators[1].Id != granted.Id {
		t.Fatalf("administrators = %+v", administrators)
	}

	// Being an administrator gives no access to procedures.
	res, err = get(t, signedIn(t, srv.URL, "titular@fi.uba.ar"), srv.URL+"/me")
	var user api.User
	decode(t, res, err, http.StatusOK, &user)
	if !user.Administrator || user.Staff {
		t.Fatalf("user = %+v, want an administrator who is not staff", user)
	}

	res, err = remove(t, admin, srv.URL+"/administrators/A-u3")
	expectStatus(t, res, err, http.StatusNoContent)
	res, err = get(t, admin, srv.URL+"/administrators")
	expectError(t, res, err, http.StatusForbidden)
	res, err = remove(t, signedIn(t, srv.URL, "titular@fi.uba.ar"), srv.URL+"/administrators/A-u3")
	expectError(t, res, err, http.StatusNotFound)
}
