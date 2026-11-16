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
	Next        string
	Referral    *ReferralRecord
	Returned    bool
}

type ReferralRecord struct {
	From               string
	To                 string
	ExternalFileNumber string
}
