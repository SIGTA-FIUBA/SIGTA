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
