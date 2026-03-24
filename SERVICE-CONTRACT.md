// @observer-project: observer
// @observer-path: SERVICE-CONTRACT.md
# SERVICE-CONTRACT.md — Observer
# @version: 0.2.0-phase2
# @updated: 2026-03-25

**Port:** 8086 · **Domain:** Observer (read-only)

---

## Code

```
internal/collector/nexus.go    polls GET /events?since=<id> every 5s — trace ID discovery
internal/trace/store.go        ring buffer, 200 entries, sync.RWMutex
internal/api/handler/traces.go GET /traces/recent · GET /traces/:trace_id
```

---

## Contract

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | none | Liveness |
| GET | `/traces/recent` | token | Last 200 known trace IDs |
| GET | `/traces/:trace_id` | token | Full correlated timeline |

`GET /traces/:trace_id` queries Nexus and Forge concurrently on demand. Empty timeline returned on upstream failure.

---

## Control

Ring buffer holds last 200 trace IDs — in-memory, no persistence. `NexusCollector.lastEventID` written only by single polling goroutine. Per-request: two goroutines (Nexus + Forge) merged after both complete. Lost on restart.

---

## Context

Derives traces from Nexus and Forge. Not authoritative for events or executions. Never calls write endpoints.
