# WORKFLOW-SESSION.md
# Session: OB-phase1-observer-tracing
# Date: 2026-03-17

## What changed — Observer Phase 1 (ADR-014)

New distributed tracing service. Discovers trace IDs from Nexus events,
assembles full correlated timelines from Nexus + Forge on demand.

## Setup and run

mkdir -p ~/workspace/projects/apps/observer
cd ~/workspace/projects/apps/observer
unzip -o /mnt/c/Users/harsh/Downloads/engx-drop/observer-phase1-tracing-20260317.zip -d .
go mod tidy && go build ./...
go install ./cmd/observer/ && cp ~/go/bin/observer ~/bin/observer
OBSERVER_SERVICE_TOKEN=7d5fcbe4-44b9-4a8f-8b79-f80925c1330e observer &

## Verify

curl -s http://127.0.0.1:8086/health
curl -s http://127.0.0.1:8086/traces/recent | jq '.data.traces'
# Get a trace ID from recent, then:
# curl -s http://127.0.0.1:8086/traces/<trace_id> | jq '.data'

## Commit

git init && git add . && \
git commit -m "feat: observer tracing phase 1 (ADR-014)" && \
git tag v0.1.0-phase1
