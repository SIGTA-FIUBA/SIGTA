package rest_test

import (
	"encoding/json"
	"io"
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

// decode reads the body of a response into body, and closes it.
func decode(t *testing.T, res *http.Response, err error, want int, body any) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d", res.Request.Method, res.Request.URL.Path, res.StatusCode, want)
	}
	if err := json.Unmarshal(raw, body); err != nil {
		t.Fatal(err)
	}
}

func TestAnAdministratorCreatesAnAreaAndGrantsAMembershipInIt(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := post(t, admin, srv.URL+"/areas", strings.NewReader(`{"id": "degree_office", "name": "Títulos"}`))
	var area api.Area
	decode(t, res, err, http.StatusCreated, &area)
	if area.Id != "degree_office" || area.Name != "Títulos" || len(area.Permissions) != 0 {
		t.Fatalf("area = %+v", area)
	}

	res, err = post(t, admin, srv.URL+"/areas/degree_office/permissions", strings.NewReader(`{"id": "sign", "name": "Firmar"}`))
	var permission api.Permission
	decode(t, res, err, http.StatusCreated, &permission)
	if permission != (api.Permission{Id: "sign", Name: "Firmar"}) {
		t.Fatalf("permission = %+v", permission)
	}

	res, err = post(t, admin, srv.URL+"/memberships", strings.NewReader(`{"user_id": "u1", "area": "degree_office", "permissions": ["sign"], "scope": [2]}`))
	var membership api.Membership
	decode(t, res, err, http.StatusCreated, &membership)
	if membership.UserId != "u1" || membership.Area != "degree_office" || !slices.Equal(membership.Permissions, []string{"sign"}) ||
		len(membership.Scope) != 1 || membership.Scope[0] != 2 || membership.GrantedBy != "u3" || membership.Until != nil {
		t.Fatalf("membership = %+v", membership)
	}

	// The membership makes the applicant staff.
	res, err = get(t, signedIn(t, srv.URL, "titular@fi.uba.ar"), srv.URL+"/me")
	var user api.User
	decode(t, res, err, http.StatusOK, &user)
	if !user.Staff || user.Administrator {
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
		"reserved id":      `{"id": "secretariat", "name": "Otra"}`,
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
	res, err := post(t, admin, srv.URL+"/areas/secretariat/permissions", strings.NewReader(`{"id": "publish_procedure_types", "name": "Otro"}`))
	expectError(t, res, err, http.StatusBadRequest)
	res, err = post(t, admin, srv.URL+"/areas/dean/permissions", strings.NewReader(`{"id": "sign", "name": "Firmar"}`))
	expectError(t, res, err, http.StatusBadRequest)

	for name, tc := range map[string]struct{ body, message string }{
		"unknown area":        {`{"user_id": "u1", "area": "dean", "permissions": ["review"]}`, "No existe esa área."},
		"unknown person":      {`{"user_id": "nadie", "area": "office", "permissions": ["review"]}`, "nunca ingresó a SIGTA"},
		"unknown permission":  {`{"user_id": "u1", "area": "office", "permissions": ["sign"]}`, "No existe ese permiso en esa área."},
		"no permissions":      {`{"user_id": "u1", "area": "office", "permissions": []}`, "al menos un rol o un permiso"},
		"no roles nor perms":  {`{"user_id": "u1", "area": "office"}`, "al menos un rol o un permiso"},
		"unknown role":        {`{"user_id": "u1", "area": "office", "roles": ["clerk"]}`, "No existe ese rol en esa área."},
		"repeated membership": {`{"user_id": "u2", "area": "office", "permissions": ["approve"]}`, "ya tiene una membresía vigente o futura"},
		"end before start":    {`{"user_id": "u1", "area": "office", "permissions": ["review"], "from": "2030-01-02T00:00:00Z", "until": "2030-01-01T00:00:00Z"}`, "La fecha de fin"},
		"end in the past":     {`{"user_id": "u1", "area": "office", "permissions": ["review"], "from": "2020-01-01T00:00:00Z", "until": "2020-02-01T00:00:00Z"}`, "La fecha de fin"},
	} {
		t.Run("membership with "+name, func(t *testing.T) {
			res, err := post(t, admin, srv.URL+"/memberships", strings.NewReader(tc.body))
			if message := expectError(t, res, err, http.StatusBadRequest); !strings.Contains(message, tc.message) {
				t.Errorf("message = %q, want it to say %q", message, tc.message)
			}
		})
	}
}

func put(t *testing.T, client *http.Client, url string, body io.Reader) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPut, url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func TestAnAdministratorGroupsPermissionsInARole(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := post(t, admin, srv.URL+"/areas/office/roles", strings.NewReader(`{"id": "clerk", "name": "Mesa", "permissions": ["review"]}`))
	var role api.Role
	decode(t, res, err, http.StatusCreated, &role)
	if role.Id != "clerk" || role.Name != "Mesa" || !slices.Equal(role.Permissions, []string{"review"}) {
		t.Fatalf("role = %+v", role)
	}

	res, err = post(t, admin, srv.URL+"/memberships", strings.NewReader(`{"user_id": "u1", "area": "office", "roles": ["clerk"]}`))
	var membership api.Membership
	decode(t, res, err, http.StatusCreated, &membership)
	if !slices.Equal(membership.Roles, []string{"clerk"}) || len(membership.Permissions) != 0 {
		t.Fatalf("membership = %+v", membership)
	}

	res, err = put(t, admin, srv.URL+"/areas/office/roles/clerk/permissions", strings.NewReader(`{"permissions": ["review", "approve"]}`))
	decode(t, res, err, http.StatusOK, &role)
	if !slices.Equal(role.Permissions, []string{"review", "approve"}) {
		t.Fatalf("role = %+v", role)
	}
}

func TestInvalidRolesAreRejected(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")
	operative := signedIn(t, srv.URL, "operativo@fi.uba.ar")

	res, err := post(t, operative, srv.URL+"/areas/office/roles", strings.NewReader(`{"id": "clerk", "name": "Mesa", "permissions": ["review"]}`))
	expectError(t, res, err, http.StatusForbidden)

	for name, body := range map[string]string{
		"no permissions":     `{"id": "clerk", "name": "Mesa", "permissions": []}`,
		"unknown permission": `{"id": "clerk", "name": "Mesa", "permissions": ["sign"]}`,
		"uppercase id":       `{"id": "Clerk", "name": "Mesa", "permissions": ["review"]}`,
	} {
		t.Run("role with "+name, func(t *testing.T) {
			res, err := post(t, admin, srv.URL+"/areas/office/roles", strings.NewReader(body))
			expectError(t, res, err, http.StatusBadRequest)
		})
	}

	res, err = put(t, admin, srv.URL+"/areas/office/roles/clerk/permissions", strings.NewReader(`{"permissions": ["review"]}`))
	expectError(t, res, err, http.StatusNotFound)
}
