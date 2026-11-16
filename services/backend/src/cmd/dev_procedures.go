package main

import (
	"context"
	"fmt"
	"time"

	proceduredomain "backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

// addDevInbox leaves made-up procedures of the example type at different
// steps, so that the inbox has something to show in local development until
// procedures can be started and moved from the interface. The applicants are
// the invented accounts of devPeople (ADR-005), and the dates are relative to
// now so that ages stay the same whenever the server starts.
func addDevInbox(ctx context.Context, versions output.ProcedureVersions, procedures output.ProcedureRepository, now time.Time) error {
	latest, err := versions.Latest(ctx, devProcedureType().ProcedureType)
	if err != nil {
		return err
	}
	daysAgo := func(days int) time.Time { return now.Add(-time.Duration(days) * 24 * time.Hour) }
	rows := []struct {
		step           string
		started, moved int
	}{
		{"review", 2, 2},
		{"review", 6, 5},
		{"review", 41, 3},
		{devDocumentsStep, 9, 4},
		{"request", 0, 0},
		{"review", 73, 6},
	}
	for i, row := range rows {
		err := procedures.Add(ctx, proceduredomain.Procedure{
			ID:            fmt.Sprintf("dev-%02d", i+1),
			ProcedureType: latest.ProcedureTypeID,
			Version:       latest.Number,
			VersionDigest: latest.Digest,
			Applicant:     fmt.Sprintf("dev-applicant-%02d", i+1),
			OpenStep:      row.step,
			StartedAt:     daysAgo(row.started),
			MovedAt:       daysAgo(row.moved),
		}, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
