package domain

// Procedure is a procedure type started by an applicant. It follows the
// version it started with, even after a new one is published.
//
// For now it only carries what authorization needs: starting it with its data
// and documents comes with US14, and advancing its steps with the step engine.
type Procedure struct {
	ID            string
	ProcedureType int64
	Version       int
	Applicant     string
	OpenStep      string
}
