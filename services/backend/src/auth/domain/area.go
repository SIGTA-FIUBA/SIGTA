package domain

// Area is a part of the organization that staff belong to, such as the
// Secretaría or a curricular commission. Areas are data: an administrator
// creates them, and procedure types refer to them by ID (ADR-017).
type Area struct {
	ID   string
	Name string
}
