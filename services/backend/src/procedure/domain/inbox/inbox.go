// Package inbox answers the staff's inbox: the procedures in progress that a
// person of the staff can see, filtered and sorted. Who sees and who acts is
// decided before, by the same rules as the detail of a procedure (ADR-020);
// this package only shapes the list.
package inbox

import (
	"cmp"
	"slices"
	"time"

	"backend/src/procedure/domain"
)

// Summary is a procedure in progress as the inbox shows it, already known to
// be visible to whoever asks.
type Summary struct {
	ID                string
	ProcedureType     int64
	ProcedureTypeName string
	ApplicantID       string
	StepName          string
	// Area is nil when the open step is the applicant's.
	Area  *Option[string]
	State domain.State
	// Actionable tells whether whoever asks holds the permissions the open
	// step asks for.
	Actionable  bool
	StartedAt   time.Time
	LastMovedAt time.Time
}

type View string

const (
	// ViewMine is what whoever asks can do now.
	ViewMine View = "mine"
	// ViewAll is every procedure in progress they can see.
	ViewAll View = "all"
)

type Sort string

const (
	// SortOldest puts first the procedures that started longest ago.
	SortOldest Sort = "oldest"
	// SortLastMoved puts first the procedures that moved most recently.
	SortLastMoved Sort = "last_moved"
)

// Query is what the inbox is asked for. A zero filter does not filter.
type Query struct {
	View          View
	ProcedureType int64
	State         domain.State
	Area          string
	Sort          Sort
}

type Item struct {
	Summary
	AgeDays int
}

type Option[T comparable] struct {
	ID   T
	Name string
}

// List holds the items that match the query, and the options to filter by,
// taken from the whole view before filtering, so that choosing a filter does
// not hide the others.
type List struct {
	Items          []Item
	ProcedureTypes []Option[int64]
	Areas          []Option[string]
}

// Build answers a query over the summaries at a given moment.
func Build(summaries []Summary, q Query, now time.Time) List {
	var inView []Item
	for _, s := range summaries {
		if q.View == ViewAll || s.Actionable {
			inView = append(inView, Item{Summary: s, AgeDays: daysBetween(s.StartedAt, now)})
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
		(q.Area == "" || item.Area != nil && item.Area.ID == q.Area)
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
	options := []Option[int64]{}
	for _, item := range items {
		option := Option[int64]{ID: item.ProcedureType, Name: item.ProcedureTypeName}
		if !slices.Contains(options, option) {
			options = append(options, option)
		}
	}
	slices.SortFunc(options, func(a, b Option[int64]) int { return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.ID, b.ID)) })
	return options
}

func areasOf(items []Item) []Option[string] {
	options := []Option[string]{}
	for _, item := range items {
		if item.Area != nil && !slices.Contains(options, *item.Area) {
			options = append(options, *item.Area)
		}
	}
	slices.SortFunc(options, func(a, b Option[string]) int { return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.ID, b.ID)) })
	return options
}
