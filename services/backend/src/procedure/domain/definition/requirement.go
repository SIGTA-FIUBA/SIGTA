package definition

// FieldRequirement is a field that a step asks for, mandatory or optional.
type FieldRequirement struct {
	Field    Field
	Required bool
}

// DocumentRequirement is a document that a step asks for, mandatory or
// optional.
type DocumentRequirement struct {
	Document Document
	Required bool
}
