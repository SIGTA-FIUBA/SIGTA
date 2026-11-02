package procedure

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func testAreas() []Area {
	return []Area{
		{ID: "applicant", Name: "Titular", Kind: AreaApplicant},
		{ID: "office", Name: "Dirección de Alumnos", Kind: AreaOffice},
		{ID: "commission", Name: "Comisión Curricular", Kind: AreaExternal},
	}
}

func testDefinition() Definition {
	return Definition{
		Name:        "Trámite de prueba",
		Description: "Un trámite para las pruebas",
		Fields:      []Field{{ID: "national_id", Name: "DNI", Type: FieldText}},
		Documents:   []Document{{ID: "letter", Name: "Carta", AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024}},
		Steps: []Step{
			{ID: "request", Name: "Solicitud", Area: "applicant", Action: ActionEnterData, Fields: []string{"national_id"}, Next: "letter"},
			{ID: "letter", Name: "Carta", Area: "applicant", Action: ActionUploadDocument, Documents: []string{"letter"}, Next: "review"},
			{ID: "review", Name: "Revisión", Area: "office", Profile: ProfileStaffOperative, Action: ActionDecide, Outcomes: []Outcome{
				{ID: "complete", Name: "Completa", Next: "closed"},
				{ID: "incomplete", Name: "Incompleta", Next: "letter"},
			}},
			{ID: "closed", Name: "Archivado", Area: "office", Profile: ProfileStaffOperative, Action: ActionConfirmExternalFact},
		},
	}
}

func messages(problems []Problem) []string {
	var got []string
	for _, p := range problems {
		got = append(got, p.String())
	}
	slices.Sort(got)
	return got
}

func TestValidDefinitions(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Definition)
	}{
		{"base definition", func(*Definition) {}},
		{"long text field", func(d *Definition) { d.Fields[0].Type = FieldLongText }},
		{"number field", func(d *Definition) { d.Fields[0].Type = FieldNumber }},
		{"date field", func(d *Definition) { d.Fields[0].Type = FieldDate }},
		{"degree field", func(d *Definition) { d.Fields[0].Type = FieldDegree }},
		{"subject list field", func(d *Definition) { d.Fields[0].Type = FieldSubjectList }},
		{"option field", func(d *Definition) {
			d.Fields[0].Type = FieldOption
			d.Fields[0].Options = []Option{{ID: "yes", Name: "Sí"}, {ID: "no", Name: "No"}}
		}},
		{"external area with approver", func(d *Definition) {
			d.Steps[3].Area = "commission"
			d.Steps[3].Profile = ProfileStaffApprover
		}},
		{"confirmation with fields", func(d *Definition) { d.Steps[3].Fields = []string{"national_id"} }},
		{"only optional fields", func(d *Definition) {
			d.Steps[0].OptionalFields = d.Steps[0].Fields
			d.Steps[0].Fields = nil
		}},
		{"only optional documents", func(d *Definition) {
			d.Steps[1].OptionalDocuments = d.Steps[1].Documents
			d.Steps[1].Documents = nil
		}},
		{"decision with its grounds", func(d *Definition) {
			d.Documents = append(d.Documents, Document{ID: "minutes", Name: "Acta", AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024})
			d.Steps[2].Documents = []string{"minutes"}
			d.Steps[2].OptionalFields = []string{"national_id"}
		}},
		{"id with digits", func(d *Definition) {
			d.Steps[3].ID = "closed2"
			d.Steps[2].Outcomes[0].Next = "closed2"
		}},
		{"two outcomes to the same step", func(d *Definition) { d.Steps[2].Outcomes[1].Next = "closed" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := testDefinition()
			tt.change(&d)
			if err := d.Validate(testAreas()); err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Definition)
		want   []string
	}{
		{"unknown action", func(d *Definition) { d.Steps[3].Action = "approve" }, []string{`step "closed": unknown action "approve"`}},
		{"unknown area", func(d *Definition) { d.Steps[0].Area = "dean" }, []string{`step "request": unknown area "dean"`}},
		{"unknown field", func(d *Definition) { d.Steps[0].Fields = []string{"email"} }, []string{
			`field "national_id" is not used by any step`,
			`step "request": unknown field "email"`,
		}},
		{"unknown document", func(d *Definition) { d.Steps[1].Documents = []string{"photo"} }, []string{
			`document "letter" is not used by any step`,
			`step "letter": unknown document "photo"`,
		}},
		{"unknown optional field", func(d *Definition) { d.Steps[0].OptionalFields = []string{"email"} }, []string{`step "request": unknown field "email"`}},
		{"unknown optional document", func(d *Definition) { d.Steps[1].OptionalDocuments = []string{"photo"} }, []string{`step "letter": unknown document "photo"`}},
		{"unknown next", func(d *Definition) { d.Steps[0].Next = "missing" }, []string{`step "request": unknown next step "missing"`}},
		{"outcome without next", func(d *Definition) { d.Steps[2].Outcomes[0].Next = "" }, []string{`step "review": outcome "complete": unknown next step ""`}},
		{"unknown outcome next", func(d *Definition) { d.Steps[2].Outcomes[0].Next = "missing" }, []string{`step "review": outcome "complete": unknown next step "missing"`}},
		{"duplicated step", func(d *Definition) { d.Steps[3].ID = "letter" }, []string{
			`step "letter" is declared twice`,
			`step "review": outcome "complete": unknown next step "closed"`,
		}},
		{"office step without profile", func(d *Definition) { d.Steps[3].Profile = "" }, []string{`step "closed": profile "" must be staff_operative or staff_approver`}},
		{"external step without profile", func(d *Definition) {
			d.Steps[3].Area = "commission"
			d.Steps[3].Profile = ""
		}, []string{`step "closed": profile "" must be staff_operative or staff_approver`}},
		{"unknown profile", func(d *Definition) { d.Steps[3].Profile = "boss" }, []string{`step "closed": profile "boss" must be staff_operative or staff_approver`}},
		{"applicant step with profile", func(d *Definition) { d.Steps[0].Profile = ProfileStaffApprover }, []string{`step "request": steps of the applicant have no profile`}},
		{"enter data without fields", func(d *Definition) { d.Steps[0].Fields = nil }, []string{
			`field "national_id" is not used by any step`,
			`step "request": enter_data needs at least one field`,
		}},
		{"upload without documents", func(d *Definition) { d.Steps[1].Documents = nil }, []string{
			`document "letter" is not used by any step`,
			`step "letter": upload_document needs at least one document`,
		}},
		{"decide with one outcome", func(d *Definition) { d.Steps[2].Outcomes = d.Steps[2].Outcomes[:1] }, []string{`step "review": decide needs at least two outcomes`}},
		{"decide with next", func(d *Definition) { d.Steps[2].Next = "closed" }, []string{`step "review": decide goes on through its outcomes, not next`}},
		{"outcomes outside decide", func(d *Definition) { d.Steps[3].Outcomes = d.Steps[2].Outcomes }, []string{`step "closed": only decide has outcomes`}},
		{"duplicated outcome name", func(d *Definition) { d.Steps[2].Outcomes[1].Name = "Completa" }, []string{`step "review": outcome name "Completa" is listed twice`}},
		{"outcome back to its own step", func(d *Definition) { d.Steps[2].Outcomes[1].Next = "review" }, []string{`step "review": outcome "incomplete" goes back to its own step`}},
		{"unreachable step", func(d *Definition) {
			d.Steps[0].Next = "review"
			d.Steps[2].Outcomes[1].Next = "request"
		}, []string{`step "letter" is unreachable from "request"`}},
		{"loop without exit", func(d *Definition) { d.Steps[2].Outcomes[0].Next = "request" }, []string{
			`step "closed" is unreachable from "request"`,
			`step "letter" never reaches a closing step`,
			`step "request" never reaches a closing step`,
			`step "review" never reaches a closing step`,
		}},
		{"no steps", func(d *Definition) { d.Steps = nil }, []string{"definition has no steps"}},
		{"no name", func(d *Definition) { d.Name = "" }, []string{"definition has no name"}},
		{"blank name", func(d *Definition) { d.Name = "   " }, []string{"definition has no name"}},
		{"name with surrounding spaces", func(d *Definition) { d.Name = "Trámite " }, []string{`definition name "Trámite " starts or ends with spaces`}},
		{"name with a leading space", func(d *Definition) { d.Name = " Trámite" }, []string{`definition name " Trámite" starts or ends with spaces`}},
		{"step without id", func(d *Definition) { d.Steps[3].ID = "" }, []string{
			"step has no id",
			`step "review": outcome "complete": unknown next step "closed"`,
		}},
		{"uppercase id", func(d *Definition) {
			d.Fields[0].ID = "NationalID"
			d.Steps[0].Fields = []string{"NationalID"}
		}, []string{`field id "NationalID" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{"lookalike id", func(d *Definition) {
			d.Steps[3].ID = "clоsed" // Cyrillic о
			d.Steps[2].Outcomes[0].Next = "clоsed"
		}, []string{`step id "clоsed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{"id starting with a digit", func(d *Definition) {
			d.Steps[3].ID = "2closed"
			d.Steps[2].Outcomes[0].Next = "2closed"
		}, []string{`step id "2closed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{"id starting with an underscore", func(d *Definition) {
			d.Steps[3].ID = "_closed"
			d.Steps[2].Outcomes[0].Next = "_closed"
		}, []string{`step id "_closed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{"id with a hyphen", func(d *Definition) {
			d.Steps[3].ID = "is-closed"
			d.Steps[2].Outcomes[0].Next = "is-closed"
		}, []string{`step id "is-closed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{"step without name", func(d *Definition) { d.Steps[0].Name = "" }, []string{`step "request" has no name`}},
		{"outcome without id", func(d *Definition) { d.Steps[2].Outcomes[0].ID = "" }, []string{`step "review": outcome has no id`}},
		{"outcome without name", func(d *Definition) { d.Steps[2].Outcomes[0].Name = "" }, []string{`step "review": outcome "complete" has no name`}},
		{"duplicated outcome", func(d *Definition) { d.Steps[2].Outcomes[1].ID = "complete" }, []string{`step "review": outcome "complete" is declared twice`}},
		{"duplicated field reference", func(d *Definition) { d.Steps[0].OptionalFields = []string{"national_id"} }, []string{`step "request": field "national_id" is listed twice`}},
		{"applicant decides", func(d *Definition) {
			d.Steps[2].Area = "applicant"
			d.Steps[2].Profile = ""
		}, []string{`step "review": the applicant only enters data or uploads documents`}},
		{"applicant confirms", func(d *Definition) {
			d.Steps[3].Area = "applicant"
			d.Steps[3].Profile = ""
		}, []string{`step "closed": the applicant only enters data or uploads documents`}},
		{"enter data with documents", func(d *Definition) { d.Steps[0].Documents = []string{"letter"} }, []string{`step "request": enter_data takes no documents`}},
		{"upload with fields", func(d *Definition) { d.Steps[1].Fields = []string{"national_id"} }, []string{`step "letter": upload_document takes no fields`}},
		{"confirm with documents", func(d *Definition) { d.Steps[3].Documents = []string{"letter"} }, []string{`step "closed": confirm_external_fact takes no documents`}},
		{"confirm with optional documents", func(d *Definition) { d.Steps[3].OptionalDocuments = []string{"letter"} }, []string{`step "closed": confirm_external_fact takes no documents`}},
		{"option without options", func(d *Definition) { d.Fields[0].Type = FieldOption }, []string{`field "national_id": option fields need at least one option`}},
		{"options outside option", func(d *Definition) { d.Fields[0].Options = []Option{{ID: "yes", Name: "Sí"}} }, []string{`field "national_id": only option fields have options`}},
		{"option without id", func(d *Definition) {
			d.Fields[0].Type = FieldOption
			d.Fields[0].Options = []Option{{Name: "Sí"}}
		}, []string{`field "national_id": option has no id`}},
		{"option without name", func(d *Definition) {
			d.Fields[0].Type = FieldOption
			d.Fields[0].Options = []Option{{ID: "yes", Name: " "}}
		}, []string{`field "national_id": option "yes" has no name`}},
		{"duplicated option id", func(d *Definition) {
			d.Fields[0].Type = FieldOption
			d.Fields[0].Options = []Option{{ID: "yes", Name: "Sí"}, {ID: "yes", Name: "Si"}}
		}, []string{`field "national_id": option "yes" is declared twice`}},
		{"duplicated option name", func(d *Definition) {
			d.Fields[0].Type = FieldOption
			d.Fields[0].Options = []Option{{ID: "yes", Name: "Sí"}, {ID: "si", Name: "Sí"}}
		}, []string{`field "national_id": option name "Sí" is listed twice`}},
		{"unknown field type", func(d *Definition) { d.Fields[0].Type = "image" }, []string{`field "national_id": unknown type "image"`}},
		{"duplicated field", func(d *Definition) { d.Fields = append(d.Fields, d.Fields[0]) }, []string{`field "national_id" is declared twice`}},
		{"document without name", func(d *Definition) { d.Documents[0].Name = "" }, []string{`document "letter" has no name`}},
		{"duplicated document", func(d *Definition) { d.Documents = append(d.Documents, d.Documents[0]) }, []string{`document "letter" is declared twice`}},
		{"unused field", func(d *Definition) {
			d.Fields = append(d.Fields, Field{ID: "email", Name: "Correo", Type: FieldText})
		}, []string{`field "email" is not used by any step`}},
		{"unused field named like a used document", func(d *Definition) {
			d.Fields = append(d.Fields, Field{ID: "letter", Name: "Carta", Type: FieldText})
		}, []string{`field "letter" is not used by any step`}},
		{"unused document named like a used field", func(d *Definition) {
			d.Documents = append(d.Documents, Document{ID: "national_id", Name: "Documento", AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024})
		}, []string{`document "national_id" is not used by any step`}},
		{"unused document", func(d *Definition) {
			d.Documents = append(d.Documents, Document{ID: "photo", Name: "Foto", AcceptedTypes: []string{"application/pdf"}, MaxSize: 1024})
		}, []string{`document "photo" is not used by any step`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := testDefinition()
			tt.change(&d)
			got := messages(d.problems(testAreas()))
			want := slices.Sorted(slices.Values(tt.want))
			if !slices.Equal(got, want) {
				t.Errorf("problems:\n got  %q\n want %q", got, want)
			}
			if err := d.Validate(testAreas()); !errors.Is(err, ErrInvalidDefinition) {
				t.Errorf("Validate() = %v, want ErrInvalidDefinition", err)
			}
		})
	}
}

func TestInvalidAreas(t *testing.T) {
	tests := []struct {
		name   string
		change func([]Area)
		want   string
	}{
		{"duplicated area", func(a []Area) { a[1].ID = "applicant" }, `area "applicant" is declared twice`},
		{"unknown area kind", func(a []Area) { a[1].Kind = "dean" }, `area "office": unknown kind "dean"`},
		{"duplicated area name", func(a []Area) { a[2].Name = "Dirección de Alumnos" }, `area "commission": name "Dirección de Alumnos" is already used by area "office"`},
		{"no applicant area", func(a []Area) { a[0].Kind = AreaOffice }, "areas must have one applicant area, found 0"},
		{"two applicant areas", func(a []Area) { a[1].Kind = AreaApplicant }, "areas must have one applicant area, found 2"},
		{"area without id", func(a []Area) { a[0].ID = "" }, "area has no id"},
		{"area with an invalid id", func(a []Area) { a[0].ID = "Applicant" }, `area id "Applicant" must be lowercase ASCII letters, digits and underscores, starting with a letter`},
		{"area without name", func(a []Area) { a[1].Name = "" }, `area "office" has no name`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			areas := testAreas()
			tt.change(areas)
			if got := messages(validateAreas(areas)); !slices.Equal(got, []string{tt.want}) {
				t.Errorf("validateAreas() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A definition is checked against valid areas only: otherwise a duplicated
// area would reach the content of the version.
func TestValidateChecksTheAreas(t *testing.T) {
	areas := append(testAreas(), Area{ID: "office", Name: "Mesa de Entradas", Kind: AreaOffice})
	if err := testDefinition().Validate(areas); !errors.Is(err, ErrInvalidAreas) {
		t.Errorf("Validate() = %v, want ErrInvalidAreas", err)
	}
}

func TestInvalidCatalogTexts(t *testing.T) {
	cases := map[string]struct {
		change func(*Definition)
		want   Code
	}{
		"blank description":   {func(d *Definition) { d.Description = "  " }, CodeNoDescription},
		"blank deadline":      {func(d *Definition) { d.Deadlines = []string{""} }, CodeBlankText},
		"blank admissibility": {func(d *Definition) { d.Admissibility = []string{" "} }, CodeBlankText},
		"file extension":      {func(d *Definition) { d.Documents[0].AcceptedTypes = []string{".pdf"} }, CodeInvalidAcceptedType},
		"with parameters":     {func(d *Definition) { d.Documents[0].AcceptedTypes = []string{"text/plain; charset=utf-8"} }, CodeInvalidAcceptedType},
		"uppercase type":      {func(d *Definition) { d.Documents[0].AcceptedTypes = []string{"Application/PDF"} }, CodeInvalidAcceptedType},
		"negative size":       {func(d *Definition) { d.Documents[0].MaxSize = -1 }, CodeNoMaxSize},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			d := testDefinition()
			c.change(&d)
			problems := ProblemsOf(d.Validate(testAreas()))
			if len(problems) != 1 || problems[0].Code != c.want {
				t.Errorf("problems = %v, want %s", problems, c.want)
			}
			if text := problems[0].String(); strings.Contains(text, "unknown problem") {
				t.Errorf("String() = %q", text)
			}
		})
	}
}
