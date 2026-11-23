package version_test

import (
	"errors"
	"testing"

	"backend/src/procedure/domain/version"
)

func TestAuthorship(t *testing.T) {
	if err := (version.Authorship{ProposedBy: "ana", ApprovedBy: "beto"}).Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil for two people", err)
	}
	tests := []struct {
		name string
		a    version.Authorship
		want string
	}{
		{name: "no proposer", a: version.Authorship{ApprovedBy: "beto"}, want: "a version needs who proposed it"},
		{name: "no approver", a: version.Authorship{ProposedBy: "ana"}, want: "a version needs who approved it"},
		{name: "same person", a: version.Authorship{ProposedBy: "ana", ApprovedBy: "ana"}, want: "the approver must be someone other than the proposer"},
		{name: "same person in capitals", a: version.Authorship{ProposedBy: "ana", ApprovedBy: "Ana"}, want: "the approver must be someone other than the proposer"},
		{name: "proposer with a space", a: version.Authorship{ProposedBy: "ana ", ApprovedBy: "ana"}, want: `proposer "ana " has surrounding spaces or invisible characters`},
		{name: "approver with an invisible character", a: version.Authorship{ProposedBy: "ana", ApprovedBy: "ana\u200b"}, want: `approver "ana\u200b" has surrounding spaces or invisible characters`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.a.Validate()
			if !errors.Is(err, version.ErrInvalidAuthorship) || err.Error() != "invalid authorship: "+tt.want {
				t.Errorf("Validate() = %v, want %q", err, tt.want)
			}
		})
	}
}
