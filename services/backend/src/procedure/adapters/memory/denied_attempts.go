package memory

import (
	"context"
	"slices"
	"sync"

	"backend/src/procedure/domain"
	"backend/src/procedure/ports/output"
)

var _ output.DeniedAttemptLog = (*DeniedAttemptLog)(nil)

type DeniedAttemptLog struct {
	mu       sync.Mutex
	attempts []domain.DeniedAttempt
}

func NewDeniedAttemptLog() *DeniedAttemptLog {
	return &DeniedAttemptLog{}
}

func (s *DeniedAttemptLog) Record(_ context.Context, a domain.DeniedAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts = append(s.attempts, a)
	return nil
}

func (s *DeniedAttemptLog) All() []domain.DeniedAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.attempts)
}
