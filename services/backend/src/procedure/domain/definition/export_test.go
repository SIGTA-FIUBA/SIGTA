package definition

// The tests, in package definition_test, reach what is not exported through
// this file, which only builds with them.

var ValidateAreas = validateAreas

func (d ProcedureTypeDefinition) Problems(areas []Area) []ValidationProblem { return d.problems(areas) }

func (p ValidationProblem) Text() (text string, aboutLast bool) { return p.text() }
