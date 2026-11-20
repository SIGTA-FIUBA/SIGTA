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
// each area (who did each step, observations, external file numbers, the
// values and documents of staff steps) stays with staff, but for the reason of
// a send back to one of their steps. A member of staff who is the applicant of
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

// Document leaves out who attached a document of a staff step.
func (a Applicant) Document(d domain.Document) domain.Document {
	if !a.ownStep(d.Step) {
		d.UploadedBy = ""
	}
	return d
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

// Steps leaves the fields and values of staff steps out, and their documents
// but the resolution once the procedure closes. In the steps of the applicant it
// leaves out a value or a file that staff entered for the same requirement.
func (a Applicant) Steps(steps []Step) []Step {
	seen := make([]Step, 0, len(steps))
	for _, s := range steps {
		if s.Step.Role != definition.RoleApplicant {
			s.Fields, s.Values = nil, nil
			s.Documents = slices.DeleteFunc(slices.Clone(s.Documents), func(r definition.DocumentRequirement) bool {
				return !a.closed || !r.Document.Resolution
			})
		}
		s.Values = slices.DeleteFunc(slices.Clone(s.Values), func(v domain.FieldValue) bool { return !a.ownStep(v.Step) })
		files := make([]domain.Document, 0, len(s.Files))
		for _, d := range s.Files {
			if a.Sees(d) {
				files = append(files, a.Document(d))
			}
		}
		s.Files = files
		seen = append(seen, s)
	}
	return seen
}

// History keeps every step done and when, without who did the steps of staff,
// their observations (but the reason of a send back to a step of the
// applicant) and external file numbers, and only the attachments the
// applicant sees.
func (a Applicant) History(history []domain.Event) []domain.Event {
	seen := make([]domain.Event, 0, len(history))
	for _, e := range history {
		if e.Document != nil && !a.sees(e.Step, e.Document.Requirement) {
			continue
		}
		if !a.ownStep(e.Step) {
			e.By = ""
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
