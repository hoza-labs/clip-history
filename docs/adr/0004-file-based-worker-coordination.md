# 0004: Coordinate with the background worker through files

Status: Accepted

Date: 2026-07-05

## Context

The foreground CLI needs to know whether the background worker is running and needs a simple way to request shutdown.

## Decision

Use files in the configured runtime directory for interprocess coordination:

- `monitor.json` stores the worker PID and heartbeat timestamps.
- `monitor.stop` signals the worker to stop.

## Alternatives

- Named pipes: better for richer communication, but more complexity than current commands need.
- TCP or local sockets: flexible, but introduces port/socket lifecycle concerns.
- Windows events or mutexes: appropriate for some coordination, but less transparent and harder to inspect manually.

## Rationale

File-based coordination is simple, inspectable, testable, and uses only the standard library. It fits the current needs: `status` reads a heartbeat, `stop` writes a stop signal, and the controller can detect stale workers. The tradeoff is that file coordination is coarse-grained, so we should revisit this if the worker needs richer commands or streaming data.
