// @observer-project: observer
// @observer-path: internal/collector/nexus.go
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	canon "github.com/Harshmaury/Canon/identity"
	"github.com/Harshmaury/Observer/internal/trace"
)

// NexusCollector polls Nexus events for trace discovery and trace lookup.
type NexusCollector struct {
	baseURL      string
	serviceToken string
	httpClient   *http.Client
	lastEventID  int64
}

// NewNexusCollector creates a NexusCollector.
func NewNexusCollector(baseURL, serviceToken string) *NexusCollector {
	return &NexusCollector{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// nexusEvent is the raw event shape from Nexus GET /events.
type nexusEvent struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Component string `json:"component"`
	Outcome   string `json:"outcome"`
	TraceID   string `json:"trace_id"`
	CreatedAt string `json:"created_at"`
}

// PollRecent fetches new events since lastEventID and returns them.
// Updates internal lastEventID cursor.
func (c *NexusCollector) PollRecent(ctx context.Context) []nexusEvent {
	path := fmt.Sprintf("/events?since=%d&limit=100", c.lastEventID)
	resp, err := c.get(ctx, path)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var envelope struct {
		OK   bool         `json:"ok"`
		Data []nexusEvent `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil
	}
	for _, e := range envelope.Data {
		if e.ID > c.lastEventID {
			c.lastEventID = e.ID
		}
	}
	return envelope.Data
}

// GetByTrace fetches all events for a specific trace ID.
func (c *NexusCollector) GetByTrace(ctx context.Context, traceID string) []*trace.TimelineEntry {
	resp, err := c.get(ctx, "/events?trace="+traceID)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var envelope struct {
		OK   bool         `json:"ok"`
		Data []nexusEvent `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil
	}

	entries := make([]*trace.TimelineEntry, 0, len(envelope.Data))
	for _, e := range envelope.Data {
		ts, _ := time.Parse(time.RFC3339Nano, e.CreatedAt)
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

func (c *NexusCollector) get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if c.serviceToken != "" && path != "/health" {
		req.Header.Set(canon.ServiceTokenHeader, c.serviceToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("nexus: HTTP %d for %s", resp.StatusCode, path)
	}
	return resp, nil
}
