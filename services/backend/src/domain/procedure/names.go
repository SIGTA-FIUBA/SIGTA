package procedure

import (
	"regexp"
	"strings"
)

// Ids are lowercase ASCII, so they work as keys and in URLs.
var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func validateEntry(kind Kind, id, name string, seen map[string]bool) []Problem {
	var problems []Problem
	switch {
	case id == "":
		problems = append(problems, Problem{Code: CodeNoID, Path: []Ref{{Kind: kind}}})
	case !idPattern.MatchString(id):
		problems = append(problems, Problem{Code: CodeInvalidID, Path: []Ref{{Kind: kind}}, Value: id})
	}
	problems = append(problems, validateName(name, Ref{Kind: kind, ID: id})...)
	if seen[id] {
		problems = append(problems, Problem{Code: CodeDeclaredTwice, Path: []Ref{{Kind: kind, ID: id}}})
	}
	seen[id] = true
	return problems
}

func validateName(name string, subject Ref) []Problem {
	switch {
	case strings.TrimSpace(name) == "":
		return []Problem{{Code: CodeNoName, Path: []Ref{subject}}}
	case name != strings.TrimSpace(name):
		return []Problem{{Code: CodeNameWithSpaces, Path: []Ref{subject}, Value: name}}
	}
	return nil
}
