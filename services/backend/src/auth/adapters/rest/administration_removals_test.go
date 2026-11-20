package rest_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"backend/src/api"
)

func remove(t *testing.T, client *http.Client, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodDelete, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return client.Do(req)
}

func TestAnAdministratorListsTheAreasWithTheirPermissionsAndRoles(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")
	res, err := post(t, admin, srv.URL+"/areas/office/roles", strings.NewReader(`{"id": "clerk", "name": "Mesa", "permissions": ["review"]}`))
	expectStatus(t, res, err, http.StatusCreated)

	res, err = get(t, admin, srv.URL+"/areas")
	var areas []api.Area
	decode(t, res, err, http.StatusOK, &areas)
	i := slices.IndexFunc(areas, func(a api.Area) bool { return a.Id == "office" })
	j := slices.IndexFunc(areas, func(a api.Area) bool { return a.Id == "secretariat" })
	if i < 0 || j < 0 || len(areas[i].Permissions) != 2 || len(areas[i].Roles) != 1 || areas[i].System || !areas[j].System {
		t.Fatalf("areas = %+v", areas)
	}

	res, err = get(t, signedIn(t, srv.URL, "operativo@fi.uba.ar"), srv.URL+"/areas")
	expectError(t, res, err, http.StatusForbidden)
}

func TestAnAreaWithMembershipsIsNotRemoved(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := remove(t, admin, srv.URL+"/areas/office")
	if msg := expectError(t, res, err, http.StatusConflict); !strings.Contains(msg, "membresías") {
		t.Errorf("message = %q, want it to say why", msg)
	}
	res, err = remove(t, admin, srv.URL+"/areas/secretariat")
	expectError(t, res, err, http.StatusConflict)

	res, err = remove(t, admin, srv.URL+"/memberships/m1")
	expectStatus(t, res, err, http.StatusNoContent)
	res, err = remove(t, admin, srv.URL+"/memberships/m1")
	expectError(t, res, err, http.StatusConflict)
	res, err = remove(t, admin, srv.URL+"/areas/office")
	expectStatus(t, res, err, http.StatusNoContent)
	res, err = remove(t, admin, srv.URL+"/areas/office")
	expectError(t, res, err, http.StatusNotFound)
}

func TestAHeldPermissionOrRoleIsNotRemoved(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")
	res, err := post(t, admin, srv.URL+"/areas/office/roles", strings.NewReader(`{"id": "clerk", "name": "Mesa", "permissions": ["approve"]}`))
	expectStatus(t, res, err, http.StatusCreated)

	res, err = remove(t, admin, srv.URL+"/areas/office/permissions/review")
	expectError(t, res, err, http.StatusConflict)
	res, err = remove(t, admin, srv.URL+"/areas/office/permissions/approve")
	expectError(t, res, err, http.StatusConflict)
	res, err = remove(t, admin, srv.URL+"/areas/secretariat/permissions/publish_procedure_types")
	expectError(t, res, err, http.StatusConflict)
	res, err = remove(t, admin, srv.URL+"/areas/office/permissions/sign")
	expectError(t, res, err, http.StatusNotFound)

	res, err = post(t, admin, srv.URL+"/memberships", strings.NewReader(`{"user_id": "u1", "area": "office", "roles": ["clerk"]}`))
	expectStatus(t, res, err, http.StatusCreated)
	res, err = remove(t, admin, srv.URL+"/areas/office/roles/clerk")
	expectError(t, res, err, http.StatusConflict)
	res, err = remove(t, admin, srv.URL+"/memberships/m2")
	expectStatus(t, res, err, http.StatusNoContent)
	res, err = remove(t, admin, srv.URL+"/areas/office/roles/clerk")
	expectStatus(t, res, err, http.StatusNoContent)
	res, err = remove(t, admin, srv.URL+"/areas/office/permissions/approve")
	expectStatus(t, res, err, http.StatusNoContent)
}

func TestChangingThePermissionsOfAMembershipEndsItAndOpensAnother(t *testing.T) {
	srv := newServer(t)
	admin := signedIn(t, srv.URL, "admin@fi.uba.ar")

	res, err := put(t, admin, srv.URL+"/memberships/m1/permissions", strings.NewReader(`{"roles": [], "permissions": ["review", "approve"]}`))
	var membership api.Membership
	decode(t, res, err, http.StatusOK, &membership)
	if membership.Id == "m1" || membership.UserId != "u2" || !slices.Equal(membership.Permissions, []string{"review", "approve"}) || membership.GrantedBy != "u3" {
		t.Fatalf("membership = %+v", membership)
	}

	res, err = get(t, admin, srv.URL+"/areas/office/memberships")
	var memberships []api.AreaMembership
	decode(t, res, err, http.StatusOK, &memberships)
	if len(memberships) != 1 || memberships[0].Id != membership.Id || memberships[0].Person.Email != "operativo@fi.uba.ar" || memberships[0].GrantedBy.Name != "Administración" {
		t.Fatalf("memberships = %+v", memberships)
	}

	res, err = put(t, admin, srv.URL+"/memberships/m1/permissions", strings.NewReader(`{"roles": [], "permissions": ["review"]}`))
	expectError(t, res, err, http.StatusConflict)
	res, err = put(t, admin, srv.URL+"/memberships/"+membership.Id+"/permissions", strings.NewReader(`{"roles": [], "permissions": ["review", "approve"]}`))
	expectError(t, res, err, http.StatusBadRequest)
	res, err = put(t, admin, srv.URL+"/memberships/none/permissions", strings.NewReader(`{"roles": [], "permissions": ["review"]}`))
	expectError(t, res, err, http.StatusNotFound)
	res, err = get(t, admin, srv.URL+"/areas/dean/memberships")
	expectError(t, res, err, http.StatusNotFound)
}

func TestOnlyAnAdministratorRemoves(t *testing.T) {
	srv := newServer(t)
	operative := signedIn(t, srv.URL, "operativo@fi.uba.ar")
	anonymous := newClient(t)

	for _, client := range []struct {
		name   string
		client *http.Client
		want   int
	}{{"operative", operative, http.StatusForbidden}, {"anonymous", anonymous, http.StatusUnauthorized}} {
		t.Run(client.name, func(t *testing.T) {
			for _, url := range []string{"/areas/office", "/areas/office/permissions/review", "/areas/office/roles/clerk", "/memberships/m1", "/administrators/A-u3"} {
				res, err := remove(t, client.client, srv.URL+url)
				expectError(t, res, err, client.want)
			}
			res, err := put(t, client.client, srv.URL+"/memberships/m1/permissions", strings.NewReader(`{"roles": [], "permissions": ["approve"]}`))
			expectError(t, res, err, client.want)
			for _, url := range []string{"/areas/office/memberships", "/administrators", "/people?email=titular@fi.uba.ar"} {
				res, err := get(t, client.client, srv.URL+url)
				expectError(t, res, err, client.want)
			}
		})
	}
}
