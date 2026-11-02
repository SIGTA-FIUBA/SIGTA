package domain

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestProblemsOf(t *testing.T) {
	d := testDefinition()
	d.Steps[2].Outcomes[0].Next = "missing"
	d.Steps[0].Fields = []string{"email"}
	want := []Problem{
		{Code: CodeUnusedField, Path: []Ref{{Kind: "field", ID: "national_id"}}},
		{Code: CodeUnknownField, Path: []Ref{{Kind: "step", ID: "request"}}, Value: "email"},
		{Code: CodeUnknownNextStep, Path: []Ref{{Kind: "step", ID: "review"}, {Kind: "outcome", ID: "complete"}}, Value: "missing"},
	}
	got := ProblemsOf(d.Validate(testAreas()))
	if len(got) != len(want) {
		t.Fatalf("ProblemsOf() = %v, want %v", got, want)
	}
	for _, w := range want {
		if !slices.ContainsFunc(got, func(g Problem) bool { return reflect.DeepEqual(g, w) }) {
			t.Errorf("ProblemsOf() = %v, missing %v", got, w)
		}
	}
}

func TestProblemsOfAreas(t *testing.T) {
	areas := testAreas()
	areas[2].Name = areas[1].Name
	want := Problem{Code: CodeNameInUse, Path: []Ref{{Kind: "area", ID: "commission"}}, Value: "Dirección de Alumnos", Other: "office"}
	got := ProblemsOf(testDefinition().Validate(areas))
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Errorf("ProblemsOf() = %v, want %v", got, want)
	}
}

func TestProblemsOfValidDefinition(t *testing.T) {
	if got := ProblemsOf(testDefinition().Validate(testAreas())); len(got) != 0 {
		t.Errorf("ProblemsOf() = %v, want none", got)
	}
}

// The editor reads problems as JSON, so these names are part of the contract.
func TestProblemJSON(t *testing.T) {
	p := Problem{Code: CodeUnknownNextStep, Path: []Ref{{Kind: "step", ID: "review"}, {Kind: "outcome"}}, Value: "missing"}
	got, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"code":"unknown_next_step","path":[{"kind":"step","id":"review"},{"kind":"outcome"}],"value":"missing"}`
	if string(got) != want {
		t.Errorf("json = %s, want %s", got, want)
	}
}

// codes reads the Code constants from problem.go, so that a new code cannot
// be left out of the list this test checks.
func codes(t *testing.T) []Code {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "problem.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found []Code
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if typ, ok := value.Type.(*ast.Ident); !ok || typ.Name != "Code" {
				continue
			}
			for _, expr := range value.Values {
				text, err := strconv.Unquote(expr.(*ast.BasicLit).Value)
				if err != nil {
					t.Fatal(err)
				}
				found = append(found, Code(text))
			}
		}
	}
	if len(found) == 0 {
		t.Fatal("no codes found in problem.go")
	}
	return found
}

func TestEveryCodeHasItsText(t *testing.T) {
	for _, code := range codes(t) {
		if text, _ := (Problem{Code: code}).text(); strings.HasPrefix(text, "unknown problem") {
			t.Errorf("code %q has no text", code)
		}
	}
}

func TestErrorsMatchTheirKind(t *testing.T) {
	err := testDefinition().Validate(append(testAreas(), Area{ID: "office", Name: "Otra", Kind: AreaOffice}))
	if !errors.Is(err, ErrInvalidAreas) || errors.Is(err, ErrInvalidDefinition) {
		t.Errorf("Validate() = %v, want only ErrInvalidAreas", err)
	}
	if got := ProblemsOf(errors.New("other")); got != nil {
		t.Errorf("ProblemsOf() = %v, want none for an error without problems", got)
	}
}
