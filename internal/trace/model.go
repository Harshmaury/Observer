// @observer-project: observer
// @observer-path: internal/trace/model.go
// Package trace defines the Observer tracing types.
package trace

import "time"

// TraceRef is a reference to a recently seen trace ID.
type TraceRef struct {
	TraceID    string    `json:"trace_id"`
	FirstSeen  time.Time `json:"first_seen"`
	EventCount int       `json:"event_count"`
}

// TimelineEntry is one event in a correlated trace timeline.
// Source is "nexus" or "forge". Type is the event type or "execution".
type TimelineEntry struct {
	At        time.Time `json:"at"`
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Component string    `json:"component,omitempty"`
	Outcome   string    `json:"outcome,omitempty"`
	Status    string    `json:"status,omitempty"`
	Target    string    `json:"target,omitempty"`
	Intent    string    `json:"intent,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// TraceSummary is the aggregate view of a trace.
type TraceSummary struct {
	DurationMS     int64 `json:"duration_ms"`
	EventCount     int   `json:"event_count"`
	ExecutionCount int   `json:"execution_count"`
}

// Trace is the full correlated view of a single trace ID.
type Trace struct {
	TraceID  string           `json:"trace_id"`
	Timeline []*TimelineEntry `json:"timeline"`
	Summary  TraceSummary     `json:"summary"`
}
