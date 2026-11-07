package domain

type Procedure struct {
	ID            string
	ProcedureType int64
	Version       int
	Applicant     string
	OpenStep      string
}
