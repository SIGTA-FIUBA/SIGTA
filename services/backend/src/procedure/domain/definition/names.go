package definition

import (
	"regexp"
	"strings"
)

// Ids are lowercase ASCII, so they work as keys and in URLs.
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func validateEntry(kind ElementKind, id, name string, seen map[string]bool) []ValidationProblem {
	var problems []ValidationProblem
	switch {
	case id == "":
		problems = append(problems, ValidationProblem{Code: CodeNoID, Path: []Element{{Kind: kind}}})
	case !idPattern.MatchString(id):
		problems = append(problems, ValidationProblem{Code: CodeInvalidID, Path: []Element{{Kind: kind}}, Value: id})
	}
	problems = append(problems, validateName(name, Element{Kind: kind, ID: id})...)
	if seen[id] {
		problems = append(problems, ValidationProblem{Code: CodeDeclaredTwice, Path: []Element{{Kind: kind, ID: id}}})
	}
	seen[id] = true
	return problems
}

// repeatedName reports a name already in seen. Options and outcomes are
// chosen by their names, so two with the same name could not be told apart.
func repeatedName(kind ElementKind, name string, seen map[string]bool) []ValidationProblem {
	if seen[name] {
		return []ValidationProblem{{Code: CodeRepeatedName, Path: []Element{{Kind: kind}}, Value: name}}
	}
	seen[name] = true
	return nil
}

func validateName(name string, subject Element) []ValidationProblem {
	switch {
	case strings.TrimSpace(name) == "":
		return []ValidationProblem{{Code: CodeNoName, Path: []Element{subject}}}
	case name != strings.TrimSpace(name):
		return []ValidationProblem{{Code: CodeNameWithSpaces, Path: []Element{subject}, Value: name}}
	}
	return nil
}
