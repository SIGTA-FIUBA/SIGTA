package domain

// Area is the copy of a staff area that a version keeps, so that it reads the
// same after the area is renamed. Areas belong to staff only: the steps of the
// applicant have no area (ADR-017).
type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func validateAreas(areas []Area) []Problem {
	var problems []Problem
	seen := map[string]bool{}
	names := map[string]string{}
	for _, a := range areas {
		problems = append(problems, validateEntry(KindArea, a.ID, a.Name, seen)...)
		// Areas are told apart by name on screen.
		if other, ok := names[a.Name]; ok {
			problems = append(problems, Problem{Code: CodeNameInUse, Path: []Ref{{Kind: KindArea, ID: a.ID}}, Value: a.Name, Other: other})
		} else {
			names[a.Name] = a.ID
		}
	}
	return problems
}
