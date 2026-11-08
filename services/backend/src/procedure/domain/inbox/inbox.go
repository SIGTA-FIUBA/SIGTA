// Package inbox decides what the staff sees in the procedure inbox: which
// procedures each profile can act on, the state each one is in, and how the
// list is filtered and sorted. It reads a summary of each procedure in
// progress and changes nothing.
package inbox

import (
	"cmp"
	"slices"
	"time"

	"backend/src/procedure/domain/definition"
)

// Summary is what the inbox needs of a procedure in progress: its open step,
// who has it, and when it started and reached that step.
type Summary struct {
	ID                string
	ProcedureType     int64
	ProcedureTypeName string
	ApplicantID       string
	ApplicantName     string
	StepName          string
	Area              definition.Area
	// Profile is the one the open step asks for; empty for applicant steps.
	Profile     definition.Profile
	StartedAt   time.Time
	LastMovedAt time.Time
}

// State is derived from the area of the open step. It is not the result of
// the procedure, which is not modeled (ADR-009).
type State string

const (
	StateWithApplicant State = "with_applicant"
	StateInOffice      State = "in_office"
	StateReferred      State = "referred"
)

// StateOf derives the state from the kind of area that has the open step.
func StateOf(area definition.Area) State {
	switch area.Kind {
	case definition.AreaApplicant:
		return StateWithApplicant
	case definition.AreaExternal:
		return StateReferred
	default:
		return StateInOffice
	}
}

// CanAct tells whether a profile, held in the procedure type of s, can do its
// open step. Only office steps are done by the staff, and the approver profile
// also covers the operative steps.
func CanAct(viewer definition.Profile, s Summary) bool {
	if s.Area.Kind != definition.AreaOffice {
		return false
	}
	switch viewer {
	case definition.ProfileStaffApprover:
		return s.Profile == definition.ProfileStaffOperative || s.Profile == definition.ProfileStaffApprover
	case definition.ProfileStaffOperative:
		return s.Profile == definition.ProfileStaffOperative
	}
	return false
}

type View string

const (
	// ViewMine is what the viewer's profile can act on now.
	ViewMine View = "mine"
	// ViewAll is every procedure in progress.
	ViewAll View = "all"
)

type Sort string

const (
	// SortOldest puts first the procedures that started longest ago.
	SortOldest Sort = "oldest"
	// SortLastMoved puts first the procedures that moved most recently.
	SortLastMoved Sort = "last_moved"
)

// Query is what the viewer asks for. Empty filters do not filter, and they
// combine with each other.
type Query struct {
	View          View
	ProcedureType int64
	State         State
	Area          string
	Sort          Sort
}

// Item is a procedure as the viewer sees it in the list.
type Item struct {
	Summary
	State      State
	Actionable bool
	AgeDays    int
	// DaysReferred is how long it has been with an external evaluator without
	// returning; it is set only when the state is StateReferred.
	DaysReferred *int
}

// Option is a value to filter by, with the name to show.
type Option[T comparable] struct {
	ID   T
	Name string
}

// List is the result of a query: the items, and the procedure types and
// areas present in the view before filtering, so that choosing one filter
// does not hide the options of the others.
type List struct {
	Items          []Item
	ProcedureTypes []Option[int64]
	Areas          []Option[string]
}

// Build answers a query over the procedures in progress, as seen with an
// access at a given moment. The store already returns only what the access
// can see; Build checks it again, so that a store that returns more never
// shows it.
func Build(summaries []Summary, access Access, q Query, now time.Time) List {
	var inView []Item
	for _, s := range summaries {
		if !access.CanSee(s) {
			continue
		}
		item := Item{
			Summary:    s,
			State:      StateOf(s.Area),
			Actionable: CanAct(access.ProfileIn(s.ProcedureType), s),
			AgeDays:    daysBetween(s.StartedAt, now),
		}
		if item.State == StateReferred {
			days := daysBetween(s.LastMovedAt, now)
			item.DaysReferred = &days
		}
		if q.View == ViewAll || item.Actionable {
			inView = append(inView, item)
		}
	}

	list := List{Items: []Item{}, ProcedureTypes: typesOf(inView), Areas: areasOf(inView)}
	for _, item := range inView {
		if q.matches(item) {
			list.Items = append(list.Items, item)
		}
	}
	slices.SortStableFunc(list.Items, q.compare)
	return list
}

func (q Query) matches(item Item) bool {
	return (q.ProcedureType == 0 || item.ProcedureType == q.ProcedureType) &&
		(q.State == "" || item.State == q.State) &&
		(q.Area == "" || item.Area.ID == q.Area)
}

func (q Query) compare(a, b Item) int {
	if q.Sort == SortLastMoved {
		return cmp.Or(b.LastMovedAt.Compare(a.LastMovedAt), cmp.Compare(a.ID, b.ID))
	}
	return cmp.Or(a.StartedAt.Compare(b.StartedAt), cmp.Compare(a.ID, b.ID))
}

// daysBetween counts whole days, and never less than zero.
func daysBetween(from, to time.Time) int {
	return max(0, int(to.Sub(from).Hours()/24))
}

func typesOf(items []Item) []Option[int64] {
	var options []Option[int64]
	for _, item := range items {
		option := Option[int64]{ID: item.ProcedureType, Name: item.ProcedureTypeName}
		if !slices.Contains(options, option) {
			options = append(options, option)
		}
	}
	slices.SortFunc(options, func(a, b Option[int64]) int { return cmp.Compare(a.Name, b.Name) })
	return options
}

func areasOf(items []Item) []Option[string] {
	var options []Option[string]
	for _, item := range items {
		option := Option[string]{ID: item.Area.ID, Name: item.Area.Name}
		if !slices.Contains(options, option) {
			options = append(options, option)
		}
	}
	slices.SortFunc(options, func(a, b Option[string]) int { return cmp.Compare(a.Name, b.Name) })
	return options
}
