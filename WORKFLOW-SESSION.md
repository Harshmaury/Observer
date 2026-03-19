# WORKFLOW-SESSION.md
# Session: OB-phase2-trace-buffer
# Date: 2026-03-19

## What changed — Observer Phase 2

Trace ring buffer increased from 50 to 200 entries. Busy platforms with
many concurrent workflows were evicting recent traces before they could
be queried. 200 is the right ceiling for local development platforms.

## Modified files
- internal/trace/store.go  — maxTraces 50 → 200

## Apply

cd ~/workspace/projects/apps/observer && \
unzip -o /mnt/c/Users/harsh/Downloads/engx-drop/observer-phase2-trace-buffer-20260319.zip -d . && \
go build ./...

## Verify

grep "maxTraces" internal/trace/store.go
# Expected: const maxTraces = 200

## Commit

git add internal/trace/store.go WORKFLOW-SESSION.md && \
git commit -m "feat(phase2): trace ring buffer 50 → 200" && \
git tag v0.2.0-phase2 && \
git push origin main --tags
