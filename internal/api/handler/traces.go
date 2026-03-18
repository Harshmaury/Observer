// @observer-project: observer
// @observer-path: internal/api/handler/traces.go
// TracesHandler handles GET /traces/* endpoints (ADR-014).
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/Harshmaury/Observer/internal/collector"
	"github.com/Harshmaury/Observer/internal/trace"
)

// TracesHandler handles GET /traces/recent and GET /traces/:trace_id.
type TracesHandler struct {
	store  *trace.Store
	nexus  *collector.NexusCollector
	forge  *collector.ForgeCollector
}

// NewTracesHandler creates a TracesHandler.
func NewTracesHandler(
	store *trace.Store,
	nexus *collector.NexusCollector,
	forge *collector.ForgeCollector,
) *TracesHandler {
	return &TracesHandler{store: store, nexus: nexus, forge: forge}
}

// Recent handles GET /traces/recent.
func (h *TracesHandler) Recent(w http.ResponseWriter, r *http.Request) {
	refs := h.store.Recent()
	respondOK(w, map[string]any{"traces": refs})
}

// ByID handles GET /traces/:trace_id.
// Assembles a full correlated timeline from Nexus events + Forge history.
func (h *TracesHandler) ByID(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("trace_id")
	if traceID == "" {
		respondErr(w, http.StatusBadRequest, fmt.Errorf("trace_id required"))
		return
	}

	// Context deadline is 12s — greater than the HTTP client timeout (10s)
	// so the client timeout fires before the parent context, giving clean
	// error propagation rather than silent empty results. (ISSUE-005)
	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()

	// Fetch from both sources concurrently.
	nexusCh := make(chan []*trace.TimelineEntry, 1)
	forgeCh := make(chan []*trace.TimelineEntry, 1)

	go func() { nexusCh <- h.nexus.GetByTrace(ctx, traceID) }()
	go func() { forgeCh <- h.forge.GetByTrace(ctx, traceID) }()

	nexusEntries := <-nexusCh
	forgeEntries := <-forgeCh

	// Merge and sort by time ascending.
	var timeline []*trace.TimelineEntry
	timeline = append(timeline, nexusEntries...)
	timeline = append(timeline, forgeEntries...)
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].At.Before(timeline[j].At)
	})

	if timeline == nil {
		timeline = []*trace.TimelineEntry{}
	}

	// Compute summary.
	summary := trace.TraceSummary{
		EventCount:     len(nexusEntries),
		ExecutionCount: len(forgeEntries),
	}
	if len(timeline) >= 2 {
		first := timeline[0].At
		last := timeline[len(timeline)-1].At
		summary.DurationMS = last.Sub(first).Milliseconds()
	}

	respondOK(w, &trace.Trace{
		TraceID:  traceID,
		Timeline: timeline,
		Summary:  summary,
	})
}

// ── RESPONSE HELPERS ──────────────────────────────────────────────────────────

type apiResponse struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func respondOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiResponse{OK: true, Data: data}) //nolint:errcheck
}

func respondErr(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{OK: false, Error: err.Error()}) //nolint:errcheck
}
