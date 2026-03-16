// @observer-project: observer
// @observer-path: internal/collector/forge.go
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Harshmaury/Observer/internal/trace"
)

// ForgeCollector fetches Forge execution history by trace ID.
type ForgeCollector struct {
	baseURL      string
	serviceToken string
	httpClient   *http.Client
}

// NewForgeCollector creates a ForgeCollector.
func NewForgeCollector(baseURL, serviceToken string) *ForgeCollector {
	return &ForgeCollector{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetByTrace fetches all execution records for a specific trace ID.
func (c *ForgeCollector) GetByTrace(ctx context.Context, traceID string) []*trace.TimelineEntry {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/history/%s", c.baseURL, traceID), nil)
	if err != nil {
		return nil
	}
	if c.serviceToken != "" {
		req.Header.Set("X-Service-Token", c.serviceToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return nil
	}
	defer resp.Body.Close()

	var envelope struct {
		OK   bool `json:"ok"`
		Data []struct {
			Intent     string `json:"intent"`
			Target     string `json:"target"`
			Status     string `json:"status"`
			DurationMS int64  `json:"duration_ms"`
			StartedAt  string `json:"started_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil
	}

	entries := make([]*trace.TimelineEntry, 0, len(envelope.Data))
	for _, r := range envelope.Data {
		ts, _ := time.Parse(time.RFC3339Nano, r.StartedAt)
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
