package main

import (
	"context"
	"fmt"
	"time"

	proceduredomain "backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/engine"
	"backend/src/procedure/ports/input"
	"backend/src/procedure/ports/output"
)

// addDevInbox leaves made-up procedures of the example types at different
// steps, with the history of how they got there, so that the inbox has
// something to show in local development. The applicants are the invented
// accounts of devPeople (ADR-005), and the dates are relative to now so that
// ages stay the same whenever the server starts.
func addDevInbox(ctx context.Context, versions output.ProcedureTypeVersionRepository, procedures output.ProcedureRepository, now time.Time) error {
	daysAgo := func(days int) time.Time { return now.Add(-time.Duration(days) * 24 * time.Hour) }
	rows := []struct {
		example        input.Publication
		step           string
		started, moved int
	}{
		{devProcedureType(), "review", 2, 2},
		{devProcedureType(), "review", 6, 5},
		{devProcedureType(), "review", 41, 3},
		{devProcedureType(), devDocumentsStep, 9, 4},
		{devProcedureType(), "request", 0, 0},
		{devProcedureType(), "review", 73, 6},
		{devReferralProcedureType(), devReviewStep, 1, 1},
		{devReferralProcedureType(), devEvaluateStep, 12, 5},
		{devRulingProcedureType(), devEvaluateStep, 8, 2},
	}
	for i, row := range rows {
		latest, err := versions.Latest(ctx, row.example.ProcedureType)
		if err != nil {
			return err
		}
		p := proceduredomain.Procedure{
			ID:            fmt.Sprintf("dev-%02d", i+1),
			ProcedureType: latest.ProcedureTypeID,
			Version:       latest.Number,
			VersionDigest: latest.Digest,
			Applicant:     fmt.Sprintf("dev-applicant-%02d", i%6+1),
			OpenStep:      row.step,
			StartedAt:     daysAgo(row.started),
			MovedAt:       daysAgo(row.moved),
		}
		var done []definition.Step
		for _, s := range row.example.Definition.Steps {
			if s.ID == row.step || s.Role != definition.RoleApplicant {
				break
			}
			done = append(done, s)
		}
		events := engine.Open(p, done, p.Applicant, p.StartedAt)
		if row.step == devEvaluateStep {
			p.Referral = &proceduredomain.Referral{From: "office", To: devCommitteeArea, Since: p.MovedAt}
			events = append(events, proceduredomain.Event{
				ProcedureID: p.ID,
				Step:        devReviewStep,
				Action:      definition.ActionDecide,
				By:          "dev-operative",
				At:          p.MovedAt,
				Observation: "La documentación está completa. Se deriva a la comisión.",
				Outcome:     "accept",
				Next:        devEvaluateStep,
				Referral:    &proceduredomain.ReferralRecord{From: "office", To: devCommitteeArea, ExternalFileNumber: "EX-2026-00000123"},
			})
		}
		var values []proceduredomain.FieldValue
		if row.step != "request" {
			values = devValues(p)
		}
		if err := procedures.Add(ctx, p, events, values...); err != nil {
			return err
		}
	}
	return nil
}
