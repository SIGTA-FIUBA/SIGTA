package definition

import "strconv"

type AreaKind string

const (
	AreaApplicant AreaKind = "applicant"
	AreaOffice    AreaKind = "office"
	AreaExternal  AreaKind = "external"
)

type Area struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Kind AreaKind `json:"kind"`
}

func validateAreas(areas []Area) []ValidationProblem {
	var problems []ValidationProblem
	seen := map[string]bool{}
	names := map[string]string{}
	applicants := 0
	for _, a := range areas {
		problems = append(problems, validateEntry(KindArea, a.ID, a.Name, seen)...)
		// Areas are told apart by name on screen.
		if other, ok := names[a.Name]; ok {
			problems = append(problems, ValidationProblem{Code: CodeNameInUse, Path: []Element{{Kind: KindArea, ID: a.ID}}, Value: a.Name, Other: other})
		} else {
			names[a.Name] = a.ID
		}
		switch a.Kind {
		case AreaApplicant:
			applicants++
		case AreaOffice, AreaExternal:
		default:
			problems = append(problems, ValidationProblem{Code: CodeUnknownAreaKind, Path: []Element{{Kind: KindArea, ID: a.ID}}, Value: string(a.Kind)})
		}
	}
	if applicants != 1 {
		problems = append(problems, ValidationProblem{Code: CodeApplicantAreas, Value: strconv.Itoa(applicants)})
	}
	return problems
}
