# 0004: Coordinate with the background worker through files

Status: Accepted

Date: 2026-07-05

Amended: 2026-10-31

## Context

The foreground CLI needs to know whether the background worker is running and needs a simple way to request shutdown.

## Decision

Use files in the configured runtime directory for interprocess coordination:

- `monitor.json` stores the worker PID and heartbeat timestamps.
- `monitor.stop` signals the worker to stop.

Because the foreground controller can read state while the worker writes it, and the worker's heartbeat and window callback can initiate writes concurrently, state-file access follows these guarantees:

- Serialize overlapping state access within a process.
- Write `monitor.json` to a temporary file in the runtime directory, flush and close it, then atomically replace the published file so readers never observe truncated or partial JSON.
- On Windows, retry bounded transient sharing and access conflicts around file replacement and reads.
- Stop and join the heartbeat goroutine before removing coordination files, so an in-progress heartbeat cannot recreate state during shutdown.
- Cover concurrent readers and writers and shutdown synchronization with automated tests; use the Go race detector when the environment supports it.

## Alternatives

- Named pipes: better for richer communication, but more complexity than current commands need.
- TCP or local sockets: flexible, but introduces port/socket lifecycle concerns.
- Windows events or mutexes: appropriate for some coordination, but less transparent and harder to inspect manually.
- Directly overwriting `monitor.json`: simpler, but concurrent readers can observe empty, partial, or invalid JSON, and overlapping writers can lose or corrupt state.

## Rationale

File-based coordination is simple, inspectable, testable, and uses only the standard library. It fits the current needs: `status` reads an atomically published heartbeat, `stop` writes a stop signal, and the controller can detect stale workers. Serialization, atomic replacement, bounded platform-specific retries, and joined shutdown make those operations safe when calls overlap. The tradeoff is that file coordination remains coarse-grained, so we should revisit this if the worker needs richer commands or streaming data.
