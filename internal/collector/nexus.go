// @observer-project: observer
// @observer-path: internal/collector/nexus.go
// ADR-039: Herald migration — replaces raw http.Get calls with typed herald client.
// Retry on transient failures is handled by herald (3 attempts, backoff).
// ADR-037: EventDTO now carries SpanID, ParentSpanID, Level — available in TimelineEntry.
package collector

import (
	"context"
	"time"

	"github.com/Harshmaury/Observer/internal/trace"
	herald "github.com/Harshmaury/Herald/client"
)

// NexusCollector polls Nexus events for trace discovery and trace lookup.
type NexusCollector struct {
	client      *herald.Client
	lastEventID int64
}

// NewNexusCollector creates a NexusCollector.
func NewNexusCollector(baseURL, serviceToken string) *NexusCollector {
	return &NexusCollector{
		client: herald.New(baseURL, herald.WithToken(serviceToken)),
	}
}

// PollRecent fetches new events since lastEventID and returns them as TimelineEntries.
// Updates internal lastEventID cursor.
func (c *NexusCollector) PollRecent(ctx context.Context) []nexusEvent {
	events, err := c.client.Events().Since(ctx, c.lastEventID, 100)
	if err != nil {
		return nil
	}
	result := make([]nexusEvent, 0, len(events))
	for _, e := range events {
		if e.ID > c.lastEventID {
			c.lastEventID = e.ID
		}
		result = append(result, nexusEvent{
			ID:        e.ID,
			Type:      e.Type,
			Component: e.Component,
			Outcome:   e.Outcome,
			TraceID:   e.TraceID,
			CreatedAt: e.CreatedAt,
		})
	}
	return result
}

// GetByTrace fetches all events for a specific trace ID as TimelineEntries.
func (c *NexusCollector) GetByTrace(ctx context.Context, traceID string) []*trace.TimelineEntry {
	events, err := c.client.Events().ByTrace(ctx, traceID)
	if err != nil {
		return nil
	}
	entries := make([]*trace.TimelineEntry, 0, len(events))
	for _, e := range events {
		ts, _ := time.Parse(time.RFC3339Nano, e.CreatedAt)
		if ts.IsZero() {
			ts, _ = time.Parse(time.RFC3339, e.CreatedAt)
		}
		entries = append(entries, &trace.TimelineEntry{
			At:        ts,
			Source:    "nexus",
			Type:      e.Type,
			Component: e.Component,
			Outcome:   e.Outcome,
		})
	}
	return entries
}

// nexusEvent is the internal event shape used by the Observer trace store.
// Mirrors the old raw struct — kept to avoid changing the store interface.
type nexusEvent struct {
	ID        int64
	Type      string
	Component string
	Outcome   string
	TraceID   string
	CreatedAt string
}
