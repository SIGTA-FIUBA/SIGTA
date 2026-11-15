package memory

import (
	"context"
	"maps"
	"sync"
	"time"

	"backend/src/procedure/domain/inbox"
	"backend/src/procedure/ports/output"
)

var _ output.ProcedureViews = (*ProcedureViews)(nil)

// ProcedureViews keeps when each person last saw each procedure, while the
// server runs.
type ProcedureViews struct {
	mu   sync.Mutex
	seen map[string]inbox.Seen
}

func NewProcedureViews() *ProcedureViews {
	return &ProcedureViews{seen: map[string]inbox.Seen{}}
}

func (v *ProcedureViews) SeenBy(_ context.Context, personID string) (inbox.Seen, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return maps.Clone(v.seen[personID]), nil
}

func (v *ProcedureViews) See(_ context.Context, personID, procedureID string, at time.Time) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.seen[personID] == nil {
		v.seen[personID] = inbox.Seen{}
	}
	v.seen[personID][procedureID] = at
	return nil
}
