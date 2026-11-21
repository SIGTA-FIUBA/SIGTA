// Package draft is a definition of a procedure type that is edited before it
// is published (ADR-013). Someone edits it and proposes it, and someone else
// who did not edit it approves the proposal, which publishes it as a new
// version, or rejects it, which takes it back to editing.
package draft

import (
	"errors"
	"slices"
	"strings"
	"time"

	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

type Status string

const (
	StatusEditing  Status = "editing"
	StatusProposed Status = "proposed"
)

var (
	ErrDraftProposed          = errors.New("the draft is proposed: reject it to edit it again")
	ErrDraftNotProposed       = errors.New("the draft is not proposed")
	ErrSelfApproval           = errors.New("whoever edited or proposed the draft cannot approve or reject it")
	ErrRejectionReasonMissing = errors.New("a rejection needs a reason")
)

// Rejection is the last one, kept until the draft is proposed again, so that
// whoever edits it knows what to fix.
type Rejection struct {
	By     string
	At     time.Time
	Reason string
}

// Draft is the only open draft of its procedure type. Base is the version it
// was edited from, or 0 for a procedure type that has none. Revision grows
// with every change, so that a change made on an old copy is refused instead
// of undoing the one in between. Editors are everyone who edited it, its
// creator included, in the order of their first change: none of them reviews
// the proposal, even after a rejection.
type Draft struct {
	ProcedureTypeID int64
	Base            int
	Definition      definition.ProcedureTypeDefinition
	Status          Status
	Revision        int
	EditedBy        string
	EditedAt        time.Time
	Editors         []string
	ProposedBy      string
	ProposedAt      time.Time
	Rejection       Rejection
}

func New(procedureType int64, base int, d definition.ProcedureTypeDefinition, by string, at time.Time) Draft {
	return Draft{
		ProcedureTypeID: procedureType,
		Base:            base,
		Definition:      d,
		Status:          StatusEditing,
		Revision:        1,
		EditedBy:        by,
		EditedAt:        at,
		Editors:         []string{by},
	}
}

func (d Draft) Edit(def definition.ProcedureTypeDefinition, by string, at time.Time) (Draft, error) {
	if d.Status == StatusProposed {
		return Draft{}, ErrDraftProposed
	}
	d.Definition = def
	d.EditedBy, d.EditedAt = by, at
	if !d.editedBy(by) {
		d.Editors = append(slices.Clip(d.Editors), by)
	}
	d.Revision++
	return d, nil
}

// Propose does not validate the definition: the caller does, against the
// current areas.
func (d Draft) Propose(by string, at time.Time) (Draft, error) {
	if d.Status == StatusProposed {
		return Draft{}, ErrDraftProposed
	}
	d.Status = StatusProposed
	d.ProposedBy, d.ProposedAt = by, at
	d.Revision++
	return d, nil
}

func (d Draft) Reject(by, reason string, at time.Time) (Draft, error) {
	if err := d.reviewableBy(by); err != nil {
		return Draft{}, err
	}
	if strings.TrimSpace(reason) == "" {
		return Draft{}, ErrRejectionReasonMissing
	}
	d.Status = StatusEditing
	d.Rejection = Rejection{By: by, At: at, Reason: reason}
	d.ProposedBy, d.ProposedAt = "", time.Time{}
	d.Revision++
	return d, nil
}

// Approval is the authorship of the version that approving the draft
// publishes.
func (d Draft) Approval(by string) (version.Authorship, error) {
	if err := d.reviewableBy(by); err != nil {
		return version.Authorship{}, err
	}
	return version.Authorship{ProposedBy: d.ProposedBy, ApprovedBy: by}, nil
}

// reviewableBy compares without case, like version.Authorship.Validate.
func (d Draft) reviewableBy(reviewer string) error {
	switch {
	case d.Status != StatusProposed:
		return ErrDraftNotProposed
	case strings.EqualFold(d.ProposedBy, reviewer), d.editedBy(reviewer):
		return ErrSelfApproval
	}
	return nil
}

func (d Draft) editedBy(person string) bool {
	return slices.ContainsFunc(d.Editors, func(editor string) bool {
		return strings.EqualFold(editor, person)
	})
}
