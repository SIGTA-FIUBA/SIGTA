package rest_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"backend/src/api"
)

func signedIn(t *testing.T, serverURL, email string) *http.Client {
	t.Helper()
	client := newClient(t)
	res, err := get(t, client, serverURL+"/auth/login?login_hint="+url.QueryEscape(email))
	expectStatus(t, res, err, http.StatusOK)
	return client
}

func decode[T any](t *testing.T, res *http.Response, err error, want int) T {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
	var body T
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestAnAdministratorCreatesAnAreaAndGrantsAMembershipInIt(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := post(t, admin, srv.URL+"/areas", strings.NewReader(`{"id": "degree_office", "name": "Títulos"}`))
	area := decode[api.Area](t, res, err, http.StatusCreated)
	if area.Id != "degree_office" || area.Name != "Títulos" || len(area.Permissions) != 0 {
		t.Fatalf("area = %+v", area)
	}

	res, err = post(t, admin, srv.URL+"/areas/degree_office/permissions", strings.NewReader(`{"id": "sign", "name": "Firmar"}`))
	if permission := decode[api.Permission](t, res, err, http.StatusCreated); permission != (api.Permission{Id: "sign", Name: "Firmar"}) {
		t.Fatalf("permission = %+v", permission)
	}

	res, err = post(t, admin, srv.URL+"/memberships", strings.NewReader(`{"user_id": "u1", "area": "degree_office", "permissions": ["sign"], "scope": [2]}`))
	membership := decode[api.Membership](t, res, err, http.StatusCreated)
	if membership.UserId != "u1" || membership.Area != "degree_office" || !slices.Equal(membership.Permissions, []string{"sign"}) ||
		len(membership.Scope) != 1 || membership.Scope[0] != 2 || membership.GrantedBy != "u3" || membership.Until != nil {
		t.Fatalf("membership = %+v", membership)
	}

	// The membership makes the applicant staff.
	res, err = get(t, signedIn(t, srv.URL, "titular@fi.uba.ar"), srv.URL+"/me")
	if user := decode[api.User](t, res, err, http.StatusOK); user.Role != "staff" || user.Administrator {
		t.Fatalf("user = %+v, want staff and not an administrator", user)
	}
}

func TestOnlyAnAdministratorManagesAreasAndMemberships(t *testing.T) {
	srv := newServer(t)
	operative := signedIn(t, srv.URL, "operativo@fi.uba.ar")

	res, err := post(t, operative, srv.URL+"/areas", strings.NewReader(`{"id": "degree_office", "name": "Títulos"}`))
	expectError(t, res, err, http.StatusForbidden)

	res, err = post(t, operative, srv.URL+"/areas/office/permissions", strings.NewReader(`{"id": "sign", "name": "Firmar"}`))
	expectError(t, res, err, http.StatusForbidden)

	res, err = post(t, operative, srv.URL+"/memberships", strings.NewReader(`{"user_id": "u1", "area": "office", "permissions": ["review"]}`))
	expectError(t, res, err, http.StatusForbidden)
}

func TestManagingAreasAndMembershipsNeedsASession(t *testing.T) {
	srv := newServer(t)
	client := newClient(t)

	res, err := post(t, client, srv.URL+"/areas", strings.NewReader(`{"id": "degree_office", "name": "Títulos"}`))
	expectError(t, res, err, http.StatusUnauthorized)

	res, err = post(t, client, srv.URL+"/memberships", strings.NewReader(`{"user_id": "u1", "area": "office", "permissions": ["review"]}`))
	expectError(t, res, err, http.StatusUnauthorized)
}

func TestInvalidAreasAndMembershipsAreRejected(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	for name, body := range map[string]string{
		"used id":          `{"id": "office", "name": "Otra"}`,
		"uppercase id":     `{"id": "Office", "name": "Otra"}`,
		"used name":        `{"id": "other", "name": "Secretaría"}`,
		"empty name":       `{"id": "other", "name": " "}`,
		"missing name key": `{"id": "other"}`,
	} {
		t.Run("area with "+name, func(t *testing.T) {
			res, err := post(t, admin, srv.URL+"/areas", strings.NewReader(body))
			expectError(t, res, err, http.StatusBadRequest)
		})
	}
	for name, body := range map[string]string{
		"used id":      `{"id": "review", "name": "Otro"}`,
		"used name":    `{"id": "other", "name": "Revisar"}`,
		"uppercase id": `{"id": "Sign", "name": "Firmar"}`,
	} {
		t.Run("permission with "+name, func(t *testing.T) {
			res, err := post(t, admin, srv.URL+"/areas/office/permissions", strings.NewReader(body))
			expectError(t, res, err, http.StatusBadRequest)
		})
	}
	res, err := post(t, admin, srv.URL+"/areas/dean/permissions", strings.NewReader(`{"id": "sign", "name": "Firmar"}`))
	expectError(t, res, err, http.StatusBadRequest)

	for name, body := range map[string]string{
		"unknown area":        `{"user_id": "u1", "area": "dean", "permissions": ["review"]}`,
		"unknown permission":  `{"user_id": "u1", "area": "office", "permissions": ["sign"]}`,
		"no permissions":      `{"user_id": "u1", "area": "office", "permissions": []}`,
		"repeated membership": `{"user_id": "u2", "area": "office", "permissions": ["approve"]}`,
		"end before start":    `{"user_id": "u1", "area": "office", "permissions": ["review"], "from": "2030-01-02T00:00:00Z", "until": "2030-01-01T00:00:00Z"}`,
	} {
		t.Run("membership with "+name, func(t *testing.T) {
			res, err := post(t, admin, srv.URL+"/memberships", strings.NewReader(body))
			expectError(t, res, err, http.StatusBadRequest)
		})
	}
}
