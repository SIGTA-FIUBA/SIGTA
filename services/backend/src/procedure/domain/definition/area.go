package definition

// Area is the copy of a staff area that a version keeps, with the permissions
// its steps ask for, so that it reads the same after either is renamed. Areas
// belong to staff only: the steps of the applicant have no area (ADR-020).
type Area struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type Permission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a Area) hasPermission(id string) bool {
	for _, p := range a.Permissions {
		if p.ID == id {
			return true
		}
	}
	return false
}

func validateAreas(areas []Area) []ValidationProblem {
	var problems []ValidationProblem
	seen := map[string]bool{}
	names := map[string]string{}
	for _, a := range areas {
		problems = append(problems, validateEntry(KindArea, a.ID, a.Name, seen)...)
		// Areas are told apart by name on screen.
		if other, ok := names[a.Name]; ok {
			problems = append(problems, ValidationProblem{Code: CodeNameInUse, Path: []Element{{Kind: KindArea, ID: a.ID}}, Value: a.Name, Other: other})
		} else {
			names[a.Name] = a.ID
		}
		seenPermissions := map[string]bool{}
		permissionNames := map[string]bool{}
		for _, p := range a.Permissions {
			permissionProblems := validateEntry(KindPermission, p.ID, p.Name, seenPermissions)
			permissionProblems = append(permissionProblems, repeatedName(KindPermission, p.Name, permissionNames)...)
			problems = append(problems, within(Element{Kind: KindArea, ID: a.ID}, permissionProblems)...)
		}
	}
	return problems
}
