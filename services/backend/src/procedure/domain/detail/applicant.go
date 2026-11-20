package detail

import (
	"slices"

	"backend/src/procedure/domain"
	"backend/src/procedure/domain/definition"
	"backend/src/procedure/domain/version"
)

// Applicant is what the applicant of a procedure sees of it: where it is, the
// dates of every step, what they entered and attached themselves and, once it
// closes, the documents that resolve it. What staff writes inside each area
// (observations, external file numbers, the values and documents of their
// steps) stays with staff. A member of staff who is the applicant of a
// procedure sees it as its applicant.
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
		s.Files = slices.DeleteFunc(slices.Clone(s.Files), func(d domain.Document) bool { return !a.Sees(d) })
		seen = append(seen, s)
	}
	return seen
}

// History keeps every step done, with who did it and when, without the
// observations and external file numbers of staff, and only the attachments
// the applicant sees.
func (a Applicant) History(history []domain.Event) []domain.Event {
	seen := make([]domain.Event, 0, len(history))
	for _, e := range history {
		if e.Document != nil && !a.sees(e.Step, e.Document.Requirement) {
			continue
		}
		if !a.ownStep(e.Step) {
			e.Observation = ""
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
