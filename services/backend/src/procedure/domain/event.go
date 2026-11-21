package domain

import (
	"time"

	"backend/src/procedure/domain/definition"
)

type Event struct {
	ProcedureID string
	Step        string
	Action      definition.Action
	By          string
	At          time.Time
	Observation string
	Outcome     string
	Declared    bool
	SendBack    bool
	Next        string
	Referral    *ReferralRecord
	Returned    bool
	// Document is set when the event is a document attached to the open
	// step, which does not move the procedure.
	Document *DocumentRecord
	// CorrectionRequest is set when the event asks the applicant for a
	// correction, which does not move the procedure either. Its reason is the
	// observation.
	CorrectionRequest *CorrectionRecord
	// Value is set when the event is a value the applicant entered in answer
	// to a correction, on the step of theirs that asks for it.
	Value *ValueRecord
	// CorrectionAnswer is set when the applicant sends what the pending
	// corrections asked for, which takes the procedure back to staff on the
	// same open step. It holds everything they asked for.
	CorrectionAnswer *CorrectionRecord
}

type ReferralRecord struct {
	From               string
	To                 string
	ExternalFileNumber string
}

type DocumentRecord struct {
	ID          string
	Requirement string
	FileName    string
}

type ValueRecord struct {
	Field string
}

type CorrectionRecord struct {
	Fields    []string
	Documents []string
}
