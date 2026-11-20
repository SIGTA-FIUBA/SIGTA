package detail

import (
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// Applicant is what the applicant of a procedure sees of it: where it is, the
// dates and the area of every step, what they entered and attached themselves
// and, once it closes, the documents that resolve it. What happens inside
// each area (who did each step, the outcome it chose, observations, external
// file numbers, the values and documents of staff steps) stays with staff, but
// for the reason of a send back to one of their steps. A member of staff who is the applicant of
// a procedure sees it as its applicant.
type Applicant struct {
	content version.Content
	closed  bool
}

func ForApplicant(c version.Content, p domain.Procedure) Applicant {
	return Applicant{content: c, closed: p.Closed()}
}

// Sees says whether the applicant sees a document of the procedure: those of
// their own steps and, once the procedure closes, its resolution.
func (a Applicant) Sees(d domain.Document) bool {
	return a.sees(d.Step, d.Requirement)
}

// Documents keeps the documents the applicant sees, without who attached
// those of staff steps.
func (a Applicant) Documents(attached []domain.Document) []domain.Document {
	seen := make([]domain.Document, 0, len(attached))
	for _, d := range attached {
		if !a.Sees(d) {
			continue
		}
		if !a.ownStep(d.Step) {
			d.UploadedBy = ""
		}
		seen = append(seen, d)
	}
	return seen
}

func (a Applicant) sees(step, requirement string) bool {
	if a.ownStep(step) {
		return true
	}
	return a.closed && slices.ContainsFunc(a.content.Definition.Documents, func(d definition.Document) bool {
		return d.ID == requirement && d.Resolution
	})
}

func (a Applicant) ownStep(id string) bool {
	return slices.ContainsFunc(a.content.Definition.Steps, func(s definition.Step) bool {
		return s.ID == id && s.Role == definition.RoleApplicant
	})
}

// Steps puts together the steps as the applicant sees them, with the values
// and documents they see: in their own steps, the last value or file they
// entered, even if staff entered another one for the same requirement later.
// Staff steps come without fields nor values, and without documents but the
// resolution once the procedure closes.
func (a Applicant) Steps(attached []domain.Document, entered []domain.FieldValue) []Step {
	entered = slices.DeleteFunc(slices.Clone(entered), func(v domain.FieldValue) bool { return !a.ownStep(v.Step) })
	steps := WithValues(WithFiles(Steps(a.content), a.Documents(attached)), entered)
	for i, s := range steps {
		if s.Step.Role == definition.RoleApplicant {
			continue
		}
		s.Fields, s.Values = nil, nil
		s.Documents = slices.DeleteFunc(slices.Clone(s.Documents), func(r definition.DocumentRequirement) bool {
			return !a.closed || !r.Document.Resolution
		})
		s.Files = slices.DeleteFunc(slices.Clone(s.Files), func(d domain.Document) bool {
			return !slices.ContainsFunc(s.Documents, func(r definition.DocumentRequirement) bool { return r.Document.ID == d.Requirement })
		})
		steps[i] = s
	}
	return steps
}

// History keeps every step done and when, without who did the steps of staff,
// the outcome they chose, their observations (but the reason of a send back
// to a step of the applicant) and external file numbers, and only the
// attachments the applicant sees. A send back still says it is one, and where
// it took the procedure.
func (a Applicant) History(history []domain.Event) []domain.Event {
	seen := make([]domain.Event, 0, len(history))
	for _, e := range history {
		if e.Document != nil && !a.sees(e.Step, e.Document.Requirement) {
			continue
		}
		if !a.ownStep(e.Step) {
			e.By, e.Outcome = "", ""
			// The reason of a send back is for the applicant when it takes
			// the procedure back to one of their steps (ADR-025).
			if !e.SendBack || !a.ownStep(e.Next) {
				e.Observation = ""
			}
		}
		if e.Referral != nil {
			r := *e.Referral
			r.ExternalFileNumber = ""
			e.Referral = &r
		}
		seen = append(seen, e)
	}
	return seen
}
