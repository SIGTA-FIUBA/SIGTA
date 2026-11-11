package definition_test

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/definition/definitiontest"
)

func problemsOf(err error) []definition.ValidationProblem {
	if invalid, ok := errors.AsType[*definition.ValidationError](err); ok {
		return invalid.Problems
	}
	return nil
}

func TestProblemsOf(t *testing.T) {
	d := definitiontest.Definition()
	d.Steps[2].Outcomes[0].Next = "missing"
	d.Steps[0].Fields = []string{"email"}
	want := []definition.ValidationProblem{
		{Code: definition.CodeUnusedField, Path: []definition.Element{{Kind: "field", ID: "national_id"}}},
		{Code: definition.CodeUnknownField, Path: []definition.Element{{Kind: "step", ID: "request"}}, Value: "email"},
		{Code: definition.CodeUnknownNextStep, Path: []definition.Element{{Kind: "step", ID: "review"}, {Kind: "outcome", ID: "complete"}}, Value: "missing"},
	}
	got := problemsOf(d.Validate(definitiontest.Areas()))
	if len(got) != len(want) {
		t.Fatalf("problems = %v, want %v", got, want)
	}
	for _, w := range want {
		if !slices.ContainsFunc(got, func(g definition.ValidationProblem) bool { return reflect.DeepEqual(g, w) }) {
			t.Errorf("problems = %v, missing %v", got, w)
		}
	}
}

func TestProblemsOfAreas(t *testing.T) {
	areas := definitiontest.Areas()
	areas[1].Name = areas[0].Name
	want := definition.ValidationProblem{Code: definition.CodeNameInUse, Path: []definition.Element{{Kind: "area", ID: "commission"}}, Value: "Dirección de Alumnos", Other: "office"}
	got := problemsOf(definitiontest.Definition().Validate(areas))
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Errorf("problems = %v, want %v", got, want)
	}
}

func TestProblemsOfValidDefinition(t *testing.T) {
	if got := problemsOf(definitiontest.Definition().Validate(definitiontest.Areas())); len(got) != 0 {
		t.Errorf("problems = %v, want none", got)
	}
}

// codes reads the ProblemCode constants from problem.go, so that a new code
// cannot be left out of the list this test checks.
func codes(t *testing.T) []definition.ProblemCode {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "problem.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found []definition.ProblemCode
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if typ, ok := value.Type.(*ast.Ident); !ok || typ.Name != "ProblemCode" {
				continue
			}
			for _, expr := range value.Values {
				text, err := strconv.Unquote(expr.(*ast.BasicLit).Value)
				if err != nil {
					t.Fatal(err)
				}
				found = append(found, definition.ProblemCode(text))
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
		if text, _ := (definition.ValidationProblem{Code: code}).Text(); strings.HasPrefix(text, "unknown problem") {
			t.Errorf("code %q has no text", code)
		}
	}
}

func TestErrorsMatchTheirKind(t *testing.T) {
	err := definitiontest.Definition().Validate(append(definitiontest.Areas(), definition.Area{ID: "office", Name: "Otra"}))
	if !errors.Is(err, definition.ErrInvalidAreas) || errors.Is(err, definition.ErrInvalidDefinition) {
		t.Errorf("Validate() = %v, want only ErrInvalidAreas", err)
	}
}
