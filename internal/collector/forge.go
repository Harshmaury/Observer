// @observer-project: observer
// @observer-path: internal/collector/forge.go
// ADR-039: full Herald migration — Forge history-by-trace calls now use typed client.
// Replaces: raw http.NewRequestWithContext + anonymous struct decode.
package collector

import (
	"context"
	"fmt"
	"time"

	herald "github.com/Harshmaury/Herald/client"
	"github.com/Harshmaury/Observer/internal/trace"
)

// ForgeCollector fetches Forge execution history by trace ID via Herald.
type ForgeCollector struct {
	forge *herald.Client
}

// NewForgeCollector creates a ForgeCollector.
func NewForgeCollector(baseURL, serviceToken string) *ForgeCollector {
	return &ForgeCollector{
		forge: herald.NewForService(baseURL, serviceToken),
	}
}

// GetByTrace fetches all execution records for a specific trace ID.
func (c *ForgeCollector) GetByTrace(ctx context.Context, traceID string) []*trace.TimelineEntry {
	records, err := c.forge.Forge().ByTrace(ctx, traceID)
	if err != nil {
		return nil
	}

	entries := make([]*trace.TimelineEntry, 0, len(records))
	for _, r := range records {
		ts := r.StartedAt
		if ts.IsZero() {
			ts = time.Time{}
		}
		entries = append(entries, &trace.TimelineEntry{
			At:      ts,
			Source:  "forge",
			Type:    "execution",
			Intent:  r.Intent,
			Target:  r.Target,
			Status:  r.Status,
			Message: fmt.Sprintf("%s %s in %dms", r.Intent, r.Target, r.DurationMS),
		})
	}
	return entries
}
