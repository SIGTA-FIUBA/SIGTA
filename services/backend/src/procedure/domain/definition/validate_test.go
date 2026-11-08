package definition_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/definition/definitiontest"
)

func messages(problems []definition.ValidationProblem) []string {
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
		change func(*definition.ProcedureTypeDefinition)
	}{
		{name: "base definition", change: func(*definition.ProcedureTypeDefinition) {}},
		{name: "long text field", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = definition.FieldLongText }},
		{name: "number field", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = definition.FieldNumber }},
		{name: "date field", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = definition.FieldDate }},
		{name: "degree field", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = definition.FieldDegree }},
		{name: "subject list field", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = definition.FieldSubjectList }},
		{name: "option field", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].Type = definition.FieldOption
			d.Fields[0].Options = []definition.Option{{ID: "yes", Name: "Sí"}, {ID: "no", Name: "No"}}
		}},
		{name: "external area with approver", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].Area = "commission"
			d.Steps[3].Profile = definition.ProfileStaffApprover
		}},
		{name: "confirmation with fields", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Fields = []string{"national_id"} }},
		{name: "only optional fields", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[0].OptionalFields = d.Steps[0].Fields
			d.Steps[0].Fields = nil
		}},
		{name: "only optional documents", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[1].OptionalDocuments = d.Steps[1].Documents
			d.Steps[1].Documents = nil
		}},
		{name: "decision with its grounds", change: func(d *definition.ProcedureTypeDefinition) {
			d.Documents = append(d.Documents, definition.Document{ID: "minutes", Name: "Acta", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1024})
			d.Steps[2].Documents = []string{"minutes"}
			d.Steps[2].OptionalFields = []string{"national_id"}
		}},
		{name: "id with digits", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].ID = "closed2"
			d.Steps[2].Outcomes[0].Next = "closed2"
		}},
		{name: "two outcomes to the same step", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[1].Next = "closed" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := definitiontest.Definition()
			tt.change(&d)
			if err := d.Validate(definitiontest.Areas()); err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name   string
		change func(*definition.ProcedureTypeDefinition)
		want   []string
	}{
		{name: "unknown action", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Action = "approve" }, want: []string{`step "closed": unknown action "approve"`}},
		{name: "unknown area", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Area = "dean" }, want: []string{`step "closed": unknown area "dean"`}},
		{name: "unknown field", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Fields = []string{"email"} }, want: []string{
			`field "national_id" is not used by any step`,
			`step "request": unknown field "email"`,
		}},
		{name: "unknown document", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[1].Documents = []string{"photo"} }, want: []string{
			`document "letter" is not used by any step`,
			`step "letter": unknown document "photo"`,
		}},
		{name: "unknown optional field", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].OptionalFields = []string{"email"} }, want: []string{`step "request": unknown field "email"`}},
		{name: "unknown optional document", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[1].OptionalDocuments = []string{"photo"} }, want: []string{`step "letter": unknown document "photo"`}},
		{name: "unknown next", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Next = "missing" }, want: []string{`step "request": unknown next step "missing"`}},
		{name: "outcome without next", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[0].Next = "" }, want: []string{`step "review": outcome "complete": unknown next step ""`}},
		{name: "unknown outcome next", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[0].Next = "missing" }, want: []string{`step "review": outcome "complete": unknown next step "missing"`}},
		{name: "duplicated step", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].ID = "letter" }, want: []string{
			`step "letter" is declared twice`,
			`step "review": outcome "complete": unknown next step "closed"`,
		}},
		{name: "office step without profile", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Profile = "" }, want: []string{`step "closed": profile "" must be staff_operative or staff_approver`}},
		{name: "external step without profile", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].Area = "commission"
			d.Steps[3].Profile = ""
		}, want: []string{`step "closed": profile "" must be staff_operative or staff_approver`}},
		{name: "unknown profile", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Profile = "boss" }, want: []string{`step "closed": profile "boss" must be staff_operative or staff_approver`}},
		{name: "applicant step with profile", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Profile = definition.ProfileStaffApprover }, want: []string{`step "request": steps of the applicant have no profile`}},
		{name: "applicant step with area", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Area = "office" }, want: []string{`step "request": steps of the applicant have no area, found "office"`}},
		{name: "staff step without area", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Area = "" }, want: []string{`step "closed": unknown area ""`}},
		{name: "step without role", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Role = "" }, want: []string{`step "request": role "" must be applicant or staff`}},
		{name: "unknown role", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Role = "dean" }, want: []string{`step "closed": role "dean" must be applicant or staff`}},
		{name: "enter data without fields", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Fields = nil }, want: []string{
			`field "national_id" is not used by any step`,
			`step "request": enter_data needs at least one field`,
		}},
		{name: "upload without documents", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[1].Documents = nil }, want: []string{
			`document "letter" is not used by any step`,
			`step "letter": upload_document needs at least one document`,
		}},
		{name: "decide with one outcome", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes = d.Steps[2].Outcomes[:1] }, want: []string{`step "review": decide needs at least two outcomes`}},
		{name: "decide with next", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Next = "closed" }, want: []string{`step "review": decide goes on through its outcomes, not next`}},
		{name: "outcomes outside decide", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Outcomes = d.Steps[2].Outcomes }, want: []string{`step "closed": only decide has outcomes`}},
		{name: "duplicated outcome name", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[1].Name = "Completa" }, want: []string{`step "review": outcome name "Completa" is listed twice`}},
		{name: "outcome back to its own step", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[1].Next = "review" }, want: []string{`step "review": outcome "incomplete" goes back to its own step`}},
		{name: "unreachable step", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[0].Next = "review"
			d.Steps[2].Outcomes[1].Next = "request"
		}, want: []string{`step "letter" is unreachable from "request"`}},
		{name: "loop without exit", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[0].Next = "request" }, want: []string{
			`step "closed" is unreachable from "request"`,
			`step "letter" never reaches a closing step`,
			`step "request" never reaches a closing step`,
			`step "review" never reaches a closing step`,
		}},
		{name: "no steps", change: func(d *definition.ProcedureTypeDefinition) { d.Steps = nil }, want: []string{"definition has no steps"}},
		{name: "no name", change: func(d *definition.ProcedureTypeDefinition) { d.Name = "" }, want: []string{"definition has no name"}},
		{name: "blank name", change: func(d *definition.ProcedureTypeDefinition) { d.Name = "   " }, want: []string{"definition has no name"}},
		{name: "name with surrounding spaces", change: func(d *definition.ProcedureTypeDefinition) { d.Name = "Trámite " }, want: []string{`definition name "Trámite " starts or ends with spaces`}},
		{name: "name with a leading space", change: func(d *definition.ProcedureTypeDefinition) { d.Name = " Trámite" }, want: []string{`definition name " Trámite" starts or ends with spaces`}},
		{name: "step without id", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].ID = "" }, want: []string{
			"step has no id",
			`step "review": outcome "complete": unknown next step "closed"`,
		}},
		{name: "uppercase id", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].ID = "NationalID"
			d.Steps[0].Fields = []string{"NationalID"}
		}, want: []string{`field id "NationalID" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{name: "lookalike id", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].ID = "clоsed" // Cyrillic о
			d.Steps[2].Outcomes[0].Next = "clоsed"
		}, want: []string{`step id "clоsed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{name: "id starting with a digit", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].ID = "2closed"
			d.Steps[2].Outcomes[0].Next = "2closed"
		}, want: []string{`step id "2closed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{name: "id starting with an underscore", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].ID = "_closed"
			d.Steps[2].Outcomes[0].Next = "_closed"
		}, want: []string{`step id "_closed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{name: "id with a hyphen", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].ID = "is-closed"
			d.Steps[2].Outcomes[0].Next = "is-closed"
		}, want: []string{`step id "is-closed" must be lowercase ASCII letters, digits and underscores, starting with a letter`}},
		{name: "step without name", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Name = "" }, want: []string{`step "request" has no name`}},
		{name: "outcome without id", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[0].ID = "" }, want: []string{`step "review": outcome has no id`}},
		{name: "outcome without name", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[0].Name = "" }, want: []string{`step "review": outcome "complete" has no name`}},
		{name: "duplicated outcome", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[2].Outcomes[1].ID = "complete" }, want: []string{`step "review": outcome "complete" is declared twice`}},
		{name: "duplicated field reference", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].OptionalFields = []string{"national_id"} }, want: []string{`step "request": field "national_id" is listed twice`}},
		{name: "applicant decides", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[2].Role, d.Steps[2].Area, d.Steps[2].Profile = definition.RoleApplicant, "", ""
		}, want: []string{`step "review": the applicant only enters data or uploads documents`}},
		{name: "applicant confirms", change: func(d *definition.ProcedureTypeDefinition) {
			d.Steps[3].Role, d.Steps[3].Area, d.Steps[3].Profile = definition.RoleApplicant, "", ""
		}, want: []string{`step "closed": the applicant only enters data or uploads documents`}},
		{name: "enter data with documents", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[0].Documents = []string{"letter"} }, want: []string{`step "request": enter_data takes no documents`}},
		{name: "upload with fields", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[1].Fields = []string{"national_id"} }, want: []string{`step "letter": upload_document takes no fields`}},
		{name: "confirm with documents", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].Documents = []string{"letter"} }, want: []string{`step "closed": confirm_external_fact takes no documents`}},
		{name: "confirm with optional documents", change: func(d *definition.ProcedureTypeDefinition) { d.Steps[3].OptionalDocuments = []string{"letter"} }, want: []string{`step "closed": confirm_external_fact takes no documents`}},
		{name: "option without options", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = definition.FieldOption }, want: []string{`field "national_id": option fields need at least one option`}},
		{name: "options outside option", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].Options = []definition.Option{{ID: "yes", Name: "Sí"}}
		}, want: []string{`field "national_id": only option fields have options`}},
		{name: "option without id", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].Type = definition.FieldOption
			d.Fields[0].Options = []definition.Option{{Name: "Sí"}}
		}, want: []string{`field "national_id": option has no id`}},
		{name: "option without name", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].Type = definition.FieldOption
			d.Fields[0].Options = []definition.Option{{ID: "yes", Name: " "}}
		}, want: []string{`field "national_id": option "yes" has no name`}},
		{name: "duplicated option id", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].Type = definition.FieldOption
			d.Fields[0].Options = []definition.Option{{ID: "yes", Name: "Sí"}, {ID: "yes", Name: "Si"}}
		}, want: []string{`field "national_id": option "yes" is declared twice`}},
		{name: "duplicated option name", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields[0].Type = definition.FieldOption
			d.Fields[0].Options = []definition.Option{{ID: "yes", Name: "Sí"}, {ID: "si", Name: "Sí"}}
		}, want: []string{`field "national_id": option name "Sí" is listed twice`}},
		{name: "unknown field type", change: func(d *definition.ProcedureTypeDefinition) { d.Fields[0].Type = "image" }, want: []string{`field "national_id": unknown type "image"`}},
		{name: "duplicated field", change: func(d *definition.ProcedureTypeDefinition) { d.Fields = append(d.Fields, d.Fields[0]) }, want: []string{`field "national_id" is declared twice`}},
		{name: "document without name", change: func(d *definition.ProcedureTypeDefinition) { d.Documents[0].Name = "" }, want: []string{`document "letter" has no name`}},
		{name: "duplicated document", change: func(d *definition.ProcedureTypeDefinition) { d.Documents = append(d.Documents, d.Documents[0]) }, want: []string{`document "letter" is declared twice`}},
		{name: "unused field", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields = append(d.Fields, definition.Field{ID: "email", Name: "Correo", Type: definition.FieldText})
		}, want: []string{`field "email" is not used by any step`}},
		{name: "unused field named like a used document", change: func(d *definition.ProcedureTypeDefinition) {
			d.Fields = append(d.Fields, definition.Field{ID: "letter", Name: "Carta", Type: definition.FieldText})
		}, want: []string{`field "letter" is not used by any step`}},
		{name: "unused document named like a used field", change: func(d *definition.ProcedureTypeDefinition) {
			d.Documents = append(d.Documents, definition.Document{ID: "national_id", Name: "Documento", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1024})
		}, want: []string{`document "national_id" is not used by any step`}},
		{name: "unused document", change: func(d *definition.ProcedureTypeDefinition) {
			d.Documents = append(d.Documents, definition.Document{ID: "photo", Name: "Foto", AcceptedTypes: []string{"application/pdf"}, MaxSizeBytes: 1024})
		}, want: []string{`document "photo" is not used by any step`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := definitiontest.Definition()
			tt.change(&d)
			got := messages(d.Problems(definitiontest.Areas()))
			want := slices.Sorted(slices.Values(tt.want))
			if !slices.Equal(got, want) {
				t.Errorf("problems:\n got  %q\n want %q", got, want)
			}
			if err := d.Validate(definitiontest.Areas()); !errors.Is(err, definition.ErrInvalidDefinition) {
				t.Errorf("Validate() = %v, want ErrInvalidDefinition", err)
			}
		})
	}
}

func TestInvalidAreas(t *testing.T) {
	tests := []struct {
		name   string
		change func([]definition.Area)
		want   string
	}{
		{name: "duplicated area", change: func(a []definition.Area) { a[1].ID = "office" }, want: `area "office" is declared twice`},
		{name: "duplicated area name", change: func(a []definition.Area) { a[1].Name = "Dirección de Alumnos" }, want: `area "commission": name "Dirección de Alumnos" is already used by area "office"`},
		{name: "area without id", change: func(a []definition.Area) { a[0].ID = "" }, want: "area has no id"},
		{name: "area with an invalid id", change: func(a []definition.Area) { a[0].ID = "Office" }, want: `area id "Office" must be lowercase ASCII letters, digits and underscores, starting with a letter`},
		{name: "area without name", change: func(a []definition.Area) { a[0].Name = "" }, want: `area "office" has no name`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			areas := definitiontest.Areas()
			tt.change(areas)
			if got := messages(definition.ValidateAreas(areas)); !slices.Equal(got, []string{tt.want}) {
				t.Errorf("validateAreas() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A definition is checked against valid areas only: otherwise a duplicated
// area would reach the content of the version.
func TestValidateChecksTheAreas(t *testing.T) {
	areas := append(definitiontest.Areas(), definition.Area{ID: "office", Name: "Mesa de Entradas"})
	if err := definitiontest.Definition().Validate(areas); !errors.Is(err, definition.ErrInvalidAreas) {
		t.Errorf("Validate() = %v, want ErrInvalidAreas", err)
	}
}

// An editor has to point at the exact element: the problem says which text
// of the list is blank and which document accepts the invalid type.
func TestCatalogTextProblemsSayWhichElement(t *testing.T) {
	cases := map[string]struct {
		change func(*definition.ProcedureTypeDefinition)
		want   string
	}{
		"blank deadline": {func(d *definition.ProcedureTypeDefinition) {
			d.Deadlines = []string{"30 días", " "}
		}, "deadline 2 has no text"},
		"blank admissibility": {func(d *definition.ProcedureTypeDefinition) {
			d.Admissibility = []string{""}
		}, "admissibility 1 has no text"},
		"invalid type": {func(d *definition.ProcedureTypeDefinition) {
			d.Documents[0].AcceptedTypes = []string{"image/*"}
		}, `document "letter" accepts "image/*", which is not a media type`},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			d := definitiontest.Definition()
			testCase.change(&d)
			problems := problemsOf(d.Validate(definitiontest.Areas()))
			if len(problems) != 1 || problems[0].String() != testCase.want {
				t.Errorf("problems = %v, want %q", problems, testCase.want)
			}
		})
	}
}

func TestInvalidCatalogTexts(t *testing.T) {
	cases := map[string]struct {
		change func(*definition.ProcedureTypeDefinition)
		want   definition.ProblemCode
	}{
		"blank description":   {func(d *definition.ProcedureTypeDefinition) { d.Description = "  " }, definition.CodeNoDescription},
		"blank deadline":      {func(d *definition.ProcedureTypeDefinition) { d.Deadlines = []string{""} }, definition.CodeBlankText},
		"blank admissibility": {func(d *definition.ProcedureTypeDefinition) { d.Admissibility = []string{" "} }, definition.CodeBlankText},
		"file extension":      {func(d *definition.ProcedureTypeDefinition) { d.Documents[0].AcceptedTypes = []string{".pdf"} }, definition.CodeInvalidAcceptedType},
		"with parameters": {func(d *definition.ProcedureTypeDefinition) {
			d.Documents[0].AcceptedTypes = []string{"text/plain; charset=utf-8"}
		}, definition.CodeInvalidAcceptedType},
		"uppercase type": {func(d *definition.ProcedureTypeDefinition) {
			d.Documents[0].AcceptedTypes = []string{"Application/PDF"}
		}, definition.CodeInvalidAcceptedType},
		"wildcard type": {func(d *definition.ProcedureTypeDefinition) { d.Documents[0].AcceptedTypes = []string{"image/*"} }, definition.CodeInvalidAcceptedType},
		"any type":      {func(d *definition.ProcedureTypeDefinition) { d.Documents[0].AcceptedTypes = []string{"*/*"} }, definition.CodeInvalidAcceptedType},
		"repeated type": {func(d *definition.ProcedureTypeDefinition) {
			d.Documents[0].AcceptedTypes = []string{"application/pdf", "application/pdf"}
		}, definition.CodeRepeatedAcceptedType},
		"negative size": {func(d *definition.ProcedureTypeDefinition) { d.Documents[0].MaxSizeBytes = -1 }, definition.CodeNoMaxSize},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			d := definitiontest.Definition()
			testCase.change(&d)
			problems := problemsOf(d.Validate(definitiontest.Areas()))
			if len(problems) != 1 || problems[0].Code != testCase.want {
				t.Errorf("problems = %v, want %s", problems, testCase.want)
			}
			if text := problems[0].String(); strings.Contains(text, "unknown problem") {
				t.Errorf("String() = %q", text)
			}
		})
	}
}
