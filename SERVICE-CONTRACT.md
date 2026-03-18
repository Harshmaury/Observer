# SERVICE-CONTRACT.md — Observer

**Service:** observer
**Domain:** Observer
**Port:** 8086
**ADRs:** ADR-014 (distributed tracing), ADR-020 (governance)
**Version:** 0.1.0-phase1
**Updated:** 2026-03-18

---

## Role

Distributed tracing service. Discovers X-Trace-ID values from Nexus events
and assembles correlated timelines from Nexus events and Forge execution
history on demand. Observer is read-only.

---

## Inputs

- `Nexus GET /events?since=<id>` — trace ID discovery (polled every 5s)
- `Forge GET /history/:trace_id` — timeline assembly (on demand per request)

---

## Outputs

- `GET /health`
- `GET /traces/recent` — last 50 known trace IDs
- `GET /traces/:trace_id` — full correlated timeline for one trace

---

## Dependencies

| Service | Used for              | Auth required   |
|---------|-----------------------|-----------------|
| Nexus   | Trace ID discovery    | X-Service-Token |
| Forge   | Timeline assembly     | X-Service-Token |

---

## Guarantees

- Ring buffer holds last 50 trace IDs — in-memory, no persistence.
- Timeline assembly queries Nexus and Forge concurrently per request.
- Graceful degradation — upstream unavailability returns empty timeline.

## Non-Responsibilities

- Observer never calls start/stop on Nexus.
- Observer never writes to any platform database.
- Observer is not the source of truth for events or executions —
  it assembles a derived view from Nexus and Forge.

## Data Authority

Derived, non-authoritative. Ring buffer is point-in-time. Lost on restart.

## Concurrency Model

- `trace.Store` protected by `sync.RWMutex`. Ring buffer reads return copies.
- `NexusCollector.lastEventID` is written only by the single polling goroutine.
- `GET /traces/:trace_id` spawns two goroutines (Nexus + Forge) per request,
  merges results after both complete.
