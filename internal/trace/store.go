// @observer-project: observer
// @observer-path: internal/trace/store.go
// Store holds the last N unique trace IDs seen across the platform.
// No SQLite — fully in-memory, stateless across restarts.
package trace

import (
	"sync"
	"time"
)

const maxTraces = 50

// Store is a bounded in-memory set of recently seen trace IDs.
type Store struct {
	mu     sync.RWMutex
	refs   []*TraceRef          // ordered by first seen, newest last
	index  map[string]*TraceRef // fast lookup by trace ID
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{index: make(map[string]*TraceRef)}
}

// Record adds or updates a trace reference.
// If the trace ID already exists, increments its event count.
// If at capacity, drops the oldest entry.
func (s *Store) Record(traceID string, eventCount int) {
	if traceID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if ref, ok := s.index[traceID]; ok {
		ref.EventCount += eventCount
		return
	}

	// At capacity — evict oldest.
	if len(s.refs) >= maxTraces {
		oldest := s.refs[0]
		delete(s.index, oldest.TraceID)
		s.refs = s.refs[1:]
	}

	ref := &TraceRef{
		TraceID:    traceID,
		FirstSeen:  time.Now().UTC(),
		EventCount: eventCount,
	}
	s.refs = append(s.refs, ref)
	s.index[traceID] = ref
}

// Recent returns all stored trace references, newest first.
func (s *Store) Recent() []*TraceRef {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*TraceRef, len(s.refs))
	for i, r := range s.refs {
		out[len(s.refs)-1-i] = r
	}
	return out
}

// Has returns true if the trace ID is known.
func (s *Store) Has(traceID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.index[traceID]
	return ok
}
