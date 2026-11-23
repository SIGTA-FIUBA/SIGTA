package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"backend/src/api"
	"backend/src/auth/adapters/dev"
	authmem "backend/src/auth/adapters/memory"
	authrest "backend/src/auth/adapters/rest"
	authapp "backend/src/auth/application"
	"backend/src/procedure/adapters/disk"
	proceduremem "backend/src/procedure/adapters/memory"
	procedurerest "backend/src/procedure/adapters/rest"
	procedureapp "backend/src/procedure/application"
)

// The configuration of subject equivalence and the areas it goes through, as
// the team reviewed them for SIG-15. godog runs in this directory.
const (
	subjectEquivalenceDefinition = "testdata/subject_equivalence/definition.json"
	subjectEquivalenceAreas      = "testdata/subject_equivalence/areas.json"
)

// subjectEquivalenceWorld runs SIGTA over the in-memory adapters and the test
// identity provider, following the wiring of src/cmd/http.go. Everyone signs in with
// their own browser and does their part through the API, so nothing is
// loaded behind it.
type subjectEquivalenceWorld struct {
	handler        http.Handler
	users          *authmem.UserRepository
	administrators *authmem.AdministratorRepository
	administrator  string // the email of the first administrator
	dir            string
	now            time.Time
	ids            int

	// browsers and people are by email. staff are those with a membership.
	browsers map[string]*http.Client
	people   map[string]string
	staff    []string

	definition    json.RawMessage
	documents     map[string]string // document id by name
	steps         map[string]string // step id by name
	stepOrder     []string
	procedureType int64
	draft         api.ProcedureTypeDraft
	published     api.PublishedVersion
	procedure     string
	// attached holds the content attached for each document, by id.
	attached map[string][]byte

	// The result of the last request.
	status int
	body   string
}

func registerSubjectEquivalenceSteps(sc *godog.ScenarioContext) {
	w := &subjectEquivalenceWorld{
		now:      time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC),
		browsers: map[string]*http.Client{},
		people:   map[string]string{},
		attached: map[string][]byte{},
	}
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if w.dir != "" {
			return ctx, errors.Join(err, os.RemoveAll(w.dir))
		}
		return ctx, err
	})

	sc.Given(`^SIGTA starts with only the Secretariat, and "([^"]*)" signs in as its first administrator$`, w.givenFirstAdministrator)
	sc.Given(`^the administrator creates the areas of subject equivalence, with their permissions and roles$`, w.givenAreas)
	sc.Given(`^these people sign in and the administrator grants them these memberships:$`, w.givenMemberships)
	sc.Given(`^"([^"]*)" signs in, with no membership$`, w.givenApplicant)
	sc.Given(`^"([^"]*)" created the draft of subject equivalence and "([^"]*)" proposed it$`, w.givenProposed)
	sc.Given(`^subject equivalence is published$`, w.givenPublished)
	sc.Given(`^"([^"]*)" started subject equivalence and "([^"]*)" confirmed the case opening$`, w.givenCaseOpened)

	sc.When(`^"([^"]*)" creates the draft of subject equivalence with its configuration$`, w.create)
	sc.When(`^"([^"]*)" proposes the draft of subject equivalence$`, w.propose)
	sc.When(`^"([^"]*)" approves the draft of subject equivalence$`, w.approve)
	sc.When(`^"([^"]*)" starts subject equivalence with their data and their academic record$`, w.start)
	sc.When(`^the staff do each step from their inbox, in order:$`, w.doSteps)
	sc.When(`^"([^"]*)" confirms the open step$`, w.confirmOpen)
	sc.When(`^"([^"]*)" attaches the "([^"]*)"$`, w.attachNamed)

	sc.Then(`^the draft of subject equivalence has no problems with the areas$`, w.thenNoProblems)
	sc.Then(`^version 1 of subject equivalence is published, proposed by "([^"]*)" and approved by "([^"]*)"$`, w.thenPublished)
	sc.Then(`^"([^"]*)" finds "([^"]*)" in the catalog$`, w.thenInCatalog)
	sc.Then(`^subject equivalence is not published$`, w.thenNotPublished)
	sc.Then(`^the procedure has these steps, in order:$`, w.thenSteps)
	sc.Then(`^the procedure waits on "([^"]*)"$`, w.thenWaitsOn)
	sc.Then(`^only "([^"]*)" has it in their inbox to act on$`, w.thenOnlyInInboxOf)
	sc.Then(`^the confirmation is refused because they lack the permission$`, w.thenStatus(http.StatusForbidden, ""))
	sc.Then(`^the confirmation is refused because a document is missing$`, w.thenStatus(http.StatusBadRequest, "Falta adjuntar"))
	sc.Then(`^the procedure ends closed, with no open step$`, w.thenClosed)
	sc.Then(`^its history has each step once, in the order of the procedure type$`, w.thenHistory)
	sc.Then(`^"([^"]*)" downloads the "([^"]*)" that was attached$`, w.thenDownloads)
	sc.Then(`^"([^"]*)" does not see the "([^"]*)" nor the "([^"]*)"$`, w.thenDoesNotSee)
}

func (w *subjectEquivalenceWorld) newID() string {
	w.ids++
	return "id-" + strconv.Itoa(w.ids)
}

// setUp follows the start of the server in src/cmd, in memory.
func (w *subjectEquivalenceWorld) setUp() error {
	ctx := context.Background()
	now := func() time.Time { return w.now }
	files, err := os.MkdirTemp("", "sigta-subject-equivalence-")
	if err != nil {
		return err
	}
	w.dir = files
	stored, err := disk.NewFiles(files)
	if err != nil {
		return err
	}
	areas := authmem.NewAreaRepository()
	if err := authapp.EnsureSecretariat(ctx, areas); err != nil {
		return err
	}
	w.users = authmem.NewUserRepository()
	w.administrators = authmem.NewAdministratorRepository()
	memberships := authmem.NewMembershipRepository()
	procedureRepository := proceduremem.NewProcedureRepository()
	documentRepository := proceduremem.NewDocumentRepository(procedureRepository)

	areaService := authapp.NewAreaService(areas, w.administrators, memberships, now)
	membershipService := authapp.NewMembershipService(areas, w.administrators, memberships, w.users, w.newID, now)
	procedureTypes := procedureapp.NewProcedureTypeService(proceduremem.NewProcedureTypeVersionRepository(), areaService, membershipService, now)
	drafts := procedureapp.NewProcedureTypeDraftService(proceduremem.NewProcedureTypeDraftRepository(), procedureTypes, membershipService, now)
	auth := authapp.NewAuthenticatorService(
		authapp.AuthConfig{IdleTimeout: idleTimeout, AllowedDomain: "fi.uba.ar"},
		dev.NewIdentityProvider(baseURL+"/auth/callback"),
		w.users,
		authmem.NewSessionRepository(),
		now,
	)
	procedures := procedureapp.NewProcedureService(procedureTypes, procedureRepository, documentRepository, membershipService, w.newID, now)
	people := authapp.NewPeopleService(w.users, w.administrators)
	server := api.NewServer(
		authrest.NewAuthHandler(auth, membershipService, baseURL+"/", loginPage, false),
		procedurerest.NewCatalogHandler(procedureTypes),
		authrest.NewAdministrationHandler(auth, areaService, membershipService, authapp.NewAdministratorService(w.administrators, w.users, w.newID, now), people),
		procedurerest.NewDocumentHandler(auth, people, procedureapp.NewDocumentService(procedures, documentRepository, stored, w.newID, now)),
		procedurerest.NewProcedureHandler(auth, people, procedureapp.NewProcedureDetailService(procedures, documentRepository), procedureapp.NewSubmissionService(procedures, stored, w.newID, now), procedures),
		procedurerest.NewInboxHandler(auth, people, procedureapp.NewInboxService(procedures)),
		procedurerest.NewProcedureTypeDraftHandler(auth, people, drafts),
	)
	w.handler = api.Mount(http.NewServeMux(), server, slog.New(slog.DiscardHandler))
	return w.readDefinition()
}

// readDefinition keeps the configuration as it is sent, and the names of its
// steps and documents, which the scenarios use.
func (w *subjectEquivalenceWorld) readDefinition() error {
	raw, err := os.ReadFile(filepath.Clean(subjectEquivalenceDefinition))
	if err != nil {
		return err
	}
	var named struct {
		Documents []struct{ ID, Name string } `json:"documents"`
		Steps     []struct{ ID, Name string } `json:"steps"`
	}
	if err := json.Unmarshal(raw, &named); err != nil {
		return err
	}
	w.definition = raw
	w.documents, w.steps = map[string]string{}, map[string]string{}
	for _, d := range named.Documents {
		w.documents[d.Name] = d.ID
	}
	for _, s := range named.Steps {
		w.steps[s.Name] = s.ID
		w.stepOrder = append(w.stepOrder, s.ID)
	}
	return nil
}

// signIn opens a browser that goes through the identity provider and comes
// back with a session, and records who the person is.
func (w *subjectEquivalenceWorld) signIn(email string) error {
	browser, err := newBrowser(func() http.Handler { return w.handler })
	if err != nil {
		return err
	}
	w.browsers[email] = browser
	target := baseURL + "/auth/login?login_hint=" + url.QueryEscape(email)
	for range 2 {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, http.NoBody)
		if err != nil {
			return err
		}
		res, err := browser.Do(req)
		if err != nil {
			return err
		}
		if err := res.Body.Close(); err != nil {
			return err
		}
		if res.StatusCode != http.StatusFound {
			return fmt.Errorf("signing in %s: got status %d at %s", email, res.StatusCode, target)
		}
		target = res.Header.Get("Location")
	}
	var me api.User
	if err := w.call(email, http.MethodGet, "/me", nil, http.StatusOK, &me); err != nil {
		return err
	}
	w.people[email] = me.Id
	return nil
}

// send makes a request in the browser of email and records the result.
func (w *subjectEquivalenceWorld) send(email, method, path string, body io.Reader, contentType string) error {
	browser, ok := w.browsers[email]
	if !ok {
		return fmt.Errorf("%s has not signed in", email)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, baseURL+path, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := browser.Do(req)
	if err != nil {
		return err
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	w.status, w.body = res.StatusCode, string(raw)
	return res.Body.Close()
}

// call sends payload as JSON, if there is one, and decodes the answer into out
// when it has the status it wants.
func (w *subjectEquivalenceWorld) call(email, method, path string, payload any, want int, out any) error {
	var body io.Reader = http.NoBody
	contentType := ""
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body, contentType = bytes.NewReader(raw), "application/json"
	}
	if err := w.send(email, method, path, body, contentType); err != nil {
		return err
	}
	if w.status != want {
		return fmt.Errorf("%s %s by %s: got status %d with %s, want %d", method, path, email, w.status, w.body, want)
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal([]byte(w.body), out)
}

func (w *subjectEquivalenceWorld) givenFirstAdministrator(email string) error {
	if err := w.setUp(); err != nil {
		return err
	}
	if err := w.signIn(email); err != nil {
		return err
	}
	if _, err := authapp.GrantFirstAdministrator(context.Background(), w.users, w.administrators, email, w.newID, w.now); err != nil {
		return err
	}
	w.administrator = email
	var areas []api.Area
	if err := w.call(email, http.MethodGet, "/areas", nil, http.StatusOK, &areas); err != nil {
		return err
	}
	if len(areas) != 1 || !areas[0].System {
		return fmt.Errorf("SIGTA starts with %+v, want only the Secretariat", areas)
	}
	return nil
}

func (w *subjectEquivalenceWorld) givenAreas() error {
	raw, err := os.ReadFile(filepath.Clean(subjectEquivalenceAreas))
	if err != nil {
		return err
	}
	var areas []api.Area
	if err := json.Unmarshal(raw, &areas); err != nil {
		return err
	}
	for _, a := range areas {
		if err := w.call(w.administrator, http.MethodPost, "/areas", map[string]string{"id": a.Id, "name": a.Name}, http.StatusCreated, nil); err != nil {
			return err
		}
		for _, p := range a.Permissions {
			if err := w.call(w.administrator, http.MethodPost, "/areas/"+a.Id+"/permissions", map[string]string{"id": p.Id, "name": p.Name}, http.StatusCreated, nil); err != nil {
				return err
			}
		}
		for _, r := range a.Roles {
			if err := w.call(w.administrator, http.MethodPost, "/areas/"+a.Id+"/roles", api.NewRole(r), http.StatusCreated, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

// givenMemberships grants each membership as an administrator does in SIGTA:
// they look the person up and pick the area, its roles and its permissions
// by name.
func (w *subjectEquivalenceWorld) givenMemberships(table *godog.Table) error {
	var areas []api.Area
	if err := w.call(w.administrator, http.MethodGet, "/areas", nil, http.StatusOK, &areas); err != nil {
		return err
	}
	for _, row := range table.Rows[1:] {
		email, areaName := row.Cells[0].Value, row.Cells[1].Value
		if err := w.signIn(email); err != nil {
			return err
		}
		var found []api.RegisteredPerson
		if err := w.call(w.administrator, http.MethodGet, "/people?q="+url.QueryEscape(email), nil, http.StatusOK, &found); err != nil {
			return err
		}
		if len(found) != 1 {
			return fmt.Errorf("looking up %s found %+v", email, found)
		}
		i := slices.IndexFunc(areas, func(a api.Area) bool { return a.Name == areaName })
		if i < 0 {
			return fmt.Errorf("no area named %q", areaName)
		}
		area := areas[i]
		roles, err := idsByName(row.Cells[2].Value, area.Roles, func(r api.Role) (string, string) { return r.Id, r.Name })
		if err != nil {
			return err
		}
		permissions, err := idsByName(row.Cells[3].Value, area.Permissions, func(p api.Permission) (string, string) { return p.Id, p.Name })
		if err != nil {
			return err
		}
		membership := api.NewMembership{UserId: found[0].Id, Area: area.Id, Roles: &roles, Permissions: &permissions}
		if err := w.call(w.administrator, http.MethodPost, "/memberships", membership, http.StatusCreated, nil); err != nil {
			return err
		}
		w.staff = append(w.staff, email)
	}
	return nil
}

// idsByName takes a comma separated list of names and returns their ids.
func idsByName[T any](names string, options []T, named func(T) (string, string)) ([]string, error) {
	ids := []string{}
	for _, name := range list(names) {
		i := slices.IndexFunc(options, func(o T) bool { _, n := named(o); return n == name })
		if i < 0 {
			return nil, fmt.Errorf("nothing named %q", name)
		}
		id, _ := named(options[i])
		ids = append(ids, id)
	}
	return ids, nil
}

func (w *subjectEquivalenceWorld) givenApplicant(email string) error { return w.signIn(email) }

func (w *subjectEquivalenceWorld) create(email string) error {
	payload := map[string]json.RawMessage{"definition": w.definition}
	if err := w.call(email, http.MethodPost, "/procedure-types", payload, http.StatusCreated, &w.draft); err != nil {
		return err
	}
	w.procedureType = w.draft.ProcedureTypeId
	return nil
}

func (w *subjectEquivalenceWorld) draftPath(action string) string {
	return "/procedure-types/" + strconv.FormatInt(w.procedureType, 10) + "/draft/" + action
}

func (w *subjectEquivalenceWorld) propose(email string) error {
	return w.call(email, http.MethodPost, w.draftPath("proposal"), api.DraftRevision{Revision: w.draft.Revision}, http.StatusOK, &w.draft)
}

// approve records the answer: whether it is refused is up to the scenario.
func (w *subjectEquivalenceWorld) approve(email string) error {
	raw, err := json.Marshal(api.DraftRevision{Revision: w.draft.Revision})
	if err != nil {
		return err
	}
	if err := w.send(email, http.MethodPost, w.draftPath("approval"), bytes.NewReader(raw), "application/json"); err != nil {
		return err
	}
	if w.status == http.StatusOK {
		return json.Unmarshal([]byte(w.body), &w.published)
	}
	return nil
}

func (w *subjectEquivalenceWorld) givenProposed(creator, proposer string) error {
	if err := w.create(creator); err != nil {
		return err
	}
	return w.propose(proposer)
}

func (w *subjectEquivalenceWorld) givenPublished() error {
	if err := w.givenProposed("ana@fi.uba.ar", "beto@fi.uba.ar"); err != nil {
		return err
	}
	if err := w.approve("carla@fi.uba.ar"); err != nil {
		return err
	}
	return w.thenPublished("beto@fi.uba.ar", "carla@fi.uba.ar")
}

func (w *subjectEquivalenceWorld) thenNoProblems() error {
	if len(w.draft.Problems) > 0 {
		return fmt.Errorf("the draft has problems: %+v", w.draft.Problems)
	}
	return nil
}

func (w *subjectEquivalenceWorld) thenPublished(proposer, approver string) error {
	if w.status != http.StatusOK {
		return fmt.Errorf("the approval got status %d with %s", w.status, w.body)
	}
	p := w.published
	if p.ProcedureTypeId != w.procedureType || p.Number != 1 || p.ProposedBy != w.people[proposer] || p.ApprovedBy != w.people[approver] {
		return fmt.Errorf("published %+v, want version 1 proposed by %s and approved by %s", p, w.people[proposer], w.people[approver])
	}
	return nil
}

func (w *subjectEquivalenceWorld) catalog(email string) ([]api.CatalogEntry, error) {
	var entries []api.CatalogEntry
	return entries, w.call(email, http.MethodGet, "/procedure-types", nil, http.StatusOK, &entries)
}

func (w *subjectEquivalenceWorld) thenInCatalog(email, name string) error {
	entries, err := w.catalog(email)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(entries, func(e api.CatalogEntry) bool { return e.Id == w.procedureType && e.Name == name }) {
		return fmt.Errorf("the catalog has %+v, without %q", entries, name)
	}
	return nil
}

func (w *subjectEquivalenceWorld) thenNotPublished() error {
	if w.status != http.StatusForbidden || !strings.Contains(w.body, "otra persona") {
		return fmt.Errorf("the approval got status %d with %s, want it refused for another person", w.status, w.body)
	}
	entries, err := w.catalog("tomas@fi.uba.ar")
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("the catalog has %+v", entries)
	}
	return nil
}

func (w *subjectEquivalenceWorld) start(email string) error {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	values, err := json.Marshal(map[string]any{
		"student_id":         "999999",
		"degree":             "Ingeniería en Informática",
		"study_plan":         "2020",
		"entry_year":         "2024",
		"requested_subjects": []string{"Análisis Matemático II", "Física I"},
	})
	if err != nil {
		return err
	}
	if err := errors.Join(
		form.WriteField("procedure_type", strconv.FormatInt(w.procedureType, 10)),
		form.WriteField("values", string(values)),
		writeDocumentPart(form, w.documents["Historia académica descargada de SIU Guaraní"], pdf("historia académica")),
		form.Close(),
	); err != nil {
		return err
	}
	if err := w.send(email, http.MethodPost, "/procedures", &body, form.FormDataContentType()); err != nil {
		return err
	}
	if w.status != http.StatusCreated {
		return fmt.Errorf("starting the procedure got status %d with %s", w.status, w.body)
	}
	var started api.StartedProcedure
	if err := json.Unmarshal([]byte(w.body), &started); err != nil {
		return err
	}
	w.procedure = started.Id
	return nil
}

// pdf is a synthetic PDF that says what it is.
func pdf(what string) []byte {
	return []byte("%PDF-1.7\n% " + what + ", sintético\n")
}

// writeDocumentPart adds a document and its file, in the order the API reads them.
func writeDocumentPart(form *multipart.Writer, document string, content []byte) error {
	if err := form.WriteField("document", document); err != nil {
		return err
	}
	part, err := form.CreateFormFile("file", document+".pdf")
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	return err
}

func (w *subjectEquivalenceWorld) detail(email string) (api.ProcedureDetail, error) {
	var d api.ProcedureDetail
	return d, w.call(email, http.MethodGet, "/procedures/"+w.procedure, nil, http.StatusOK, &d)
}

func (w *subjectEquivalenceWorld) thenSteps(table *godog.Table) error {
	d, err := w.detail("tomas@fi.uba.ar")
	if err != nil {
		return err
	}
	var got [][]string
	for _, s := range d.Steps {
		area := ""
		if s.Area != nil {
			area = s.Area.Name
		}
		got = append(got, []string{s.Name, string(s.Role), area, string(s.Action)})
	}
	var want [][]string
	for _, row := range table.Rows[1:] {
		want = append(want, []string{row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value})
	}
	if !slices.EqualFunc(got, want, slices.Equal) {
		return fmt.Errorf("the steps are\n%v\nwant\n%v", got, want)
	}
	return nil
}

func (w *subjectEquivalenceWorld) openStep() (string, error) {
	d, err := w.detail("tomas@fi.uba.ar")
	if err != nil {
		return "", err
	}
	if d.OpenStep == nil {
		return "", nil
	}
	return *d.OpenStep, nil
}

func (w *subjectEquivalenceWorld) thenWaitsOn(step string) error {
	open, err := w.openStep()
	if err != nil {
		return err
	}
	if open != w.steps[step] {
		return fmt.Errorf("the procedure waits on %q, want %q", open, w.steps[step])
	}
	return nil
}

// inbox is what the person has to do, as their inbox shows it.
func (w *subjectEquivalenceWorld) inbox(email string) ([]api.ProcedureListItem, error) {
	var l api.ProcedureList
	return l.Items, w.call(email, http.MethodGet, "/procedures?view=mine", nil, http.StatusOK, &l)
}

func (w *subjectEquivalenceWorld) thenOnlyInInboxOf(email string) error {
	open, err := w.openStep()
	if err != nil {
		return err
	}
	for _, other := range w.staff {
		items, err := w.inbox(other)
		if err != nil {
			return err
		}
		i := slices.IndexFunc(items, func(item api.ProcedureListItem) bool { return item.Id == w.procedure })
		switch {
		case other == email && i < 0:
			return fmt.Errorf("%s does not have the procedure in their inbox", email)
		case other == email && (!items[i].Actionable || w.steps[items[i].StepName] != open):
			return fmt.Errorf("%s has %+v in their inbox, want %q to act on", email, items[i], open)
		case other != email && i >= 0:
			return fmt.Errorf("%s also has the procedure in their inbox", other)
		}
	}
	return nil
}

func (w *subjectEquivalenceWorld) attachNamed(email, document string) error {
	id, ok := w.documents[document]
	if !ok {
		return fmt.Errorf("the procedure type has no document %q", document)
	}
	content := pdf(document)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	if err := errors.Join(writeDocumentPart(form, id, content), form.Close()); err != nil {
		return err
	}
	if err := w.send(email, http.MethodPost, "/procedures/"+w.procedure+"/documents", &body, form.FormDataContentType()); err != nil {
		return err
	}
	if w.status != http.StatusCreated {
		return fmt.Errorf("%s attaching %q got status %d with %s", email, document, w.status, w.body)
	}
	w.attached[id] = content
	return nil
}

// confirm confirms the open step and records the answer.
func (w *subjectEquivalenceWorld) confirm(email, externalFileNumber string) error {
	open, err := w.openStep()
	if err != nil {
		return err
	}
	confirmation := api.StepConfirmation{Observation: "Hecho fuera de SIGTA, dato sintético"}
	if externalFileNumber != "" {
		confirmation.ExternalFileNumber = &externalFileNumber
	}
	raw, err := json.Marshal(confirmation)
	if err != nil {
		return err
	}
	return w.send(email, http.MethodPost, "/procedures/"+w.procedure+"/steps/"+open+"/confirmation", bytes.NewReader(raw), "application/json")
}

func (w *subjectEquivalenceWorld) confirmOpen(email string) error { return w.confirm(email, "") }

// do is one step done by the staff: the person finds it in their inbox,
// attaches what it asks for and confirms it.
func (w *subjectEquivalenceWorld) do(step, email, document, externalFileNumber string) error {
	if err := w.thenWaitsOn(step); err != nil {
		return err
	}
	if err := w.thenOnlyInInboxOf(email); err != nil {
		return fmt.Errorf("on %q: %w", step, err)
	}
	if document != "" {
		if err := w.attachNamed(email, document); err != nil {
			return err
		}
	}
	if err := w.confirm(email, externalFileNumber); err != nil {
		return err
	}
	if w.status != http.StatusCreated {
		return fmt.Errorf("%s confirming %q got status %d with %s", email, step, w.status, w.body)
	}
	if externalFileNumber == "" {
		return nil
	}
	return w.thenExternalFileNumber(email, step, externalFileNumber)
}

// thenExternalFileNumber looks for the number in the referral that the
// confirmation of step left in the history.
func (w *subjectEquivalenceWorld) thenExternalFileNumber(email, step, want string) error {
	d, err := w.detail(email)
	if err != nil {
		return err
	}
	for _, e := range d.History {
		if e.Step == w.steps[step] && e.Document == nil {
			if e.Referral == nil || e.Referral.ExternalFileNumber == nil || *e.Referral.ExternalFileNumber != want {
				return fmt.Errorf("%q left the referral %+v, want the external file number %s", step, e.Referral, want)
			}
			return nil
		}
	}
	return fmt.Errorf("the history has no %q", step)
}

func (w *subjectEquivalenceWorld) doSteps(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		if err := w.do(row.Cells[0].Value, row.Cells[1].Value, row.Cells[2].Value, row.Cells[3].Value); err != nil {
			return err
		}
	}
	return nil
}

func (w *subjectEquivalenceWorld) givenCaseOpened(applicant, clerk string) error {
	if err := w.start(applicant); err != nil {
		return err
	}
	return w.do("Expediente electrónico caratulado", clerk, "", "EX-2026-00000001")
}

func (w *subjectEquivalenceWorld) thenStatus(want int, says string) func() error {
	return func() error {
		if w.status != want || !strings.Contains(w.body, says) {
			return fmt.Errorf("got status %d with %s, want %d saying %q", w.status, w.body, want, says)
		}
		return nil
	}
}

func (w *subjectEquivalenceWorld) thenClosed() error {
	d, err := w.detail("tomas@fi.uba.ar")
	if err != nil {
		return err
	}
	if d.State != api.Closed || d.OpenStep != nil {
		return fmt.Errorf("the procedure is %s on %v, want it closed", d.State, d.OpenStep)
	}
	return nil
}

// thenHistory leaves out the documents attached after the start, which are
// events of the open step but not steps done.
func (w *subjectEquivalenceWorld) thenHistory() error {
	d, err := w.detail("dora@fi.uba.ar")
	if err != nil {
		return err
	}
	var done []string
	for _, e := range d.History {
		if e.Document == nil {
			done = append(done, e.Step)
		}
	}
	if !slices.Equal(done, w.stepOrder) {
		return fmt.Errorf("the history has %v, want %v", done, w.stepOrder)
	}
	return nil
}

func (w *subjectEquivalenceWorld) documentsOf(email string) ([]api.ProcedureDocument, error) {
	var documents []api.ProcedureDocument
	return documents, w.call(email, http.MethodGet, "/procedures/"+w.procedure+"/documents", nil, http.StatusOK, &documents)
}

func (w *subjectEquivalenceWorld) thenDownloads(email, document string) error {
	documents, err := w.documentsOf(email)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(documents, func(d api.ProcedureDocument) bool { return d.Document == w.documents[document] })
	if i < 0 {
		return fmt.Errorf("%s does not see %q among %+v", email, document, documents)
	}
	if err := w.call(email, http.MethodGet, "/procedures/"+w.procedure+"/documents/"+documents[i].Id+"/content", nil, http.StatusOK, nil); err != nil {
		return err
	}
	if w.body != string(w.attached[w.documents[document]]) {
		return fmt.Errorf("%s got %q, want what was attached", email, w.body)
	}
	return nil
}

// thenDoesNotSee checks the list and the download, with the ids the staff see.
func (w *subjectEquivalenceWorld) thenDoesNotSee(email, document, other string) error {
	documents := []string{document, other}
	theirs, err := w.documentsOf(email)
	if err != nil {
		return err
	}
	all, err := w.documentsOf("dora@fi.uba.ar")
	if err != nil {
		return err
	}
	for _, document := range documents {
		id := w.documents[document]
		if slices.ContainsFunc(theirs, func(d api.ProcedureDocument) bool { return d.Document == id }) {
			return fmt.Errorf("%s sees %q", email, document)
		}
		i := slices.IndexFunc(all, func(d api.ProcedureDocument) bool { return d.Document == id })
		if i < 0 {
			return fmt.Errorf("%q was not attached", document)
		}
		if err := w.send(email, http.MethodGet, "/procedures/"+w.procedure+"/documents/"+all[i].Id+"/content", http.NoBody, ""); err != nil {
			return err
		}
		if w.status != http.StatusNotFound {
			return fmt.Errorf("%s downloading %q got status %d, want 404", email, document, w.status)
		}
	}
	return nil
}
