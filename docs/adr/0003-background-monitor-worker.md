# 0003: Run clipboard monitoring as a background worker process

Status: Accepted

Date: 2026-07-05

## Context

Clipboard monitoring must continue after `ch start` exits. The foreground command should not block forever, and users need separate `stop` and `status` commands.

## Decision

Use `ch start` to launch a detached background worker process via the hidden `ch monitor` command. The monitor package owns the controller, worker, and worker process management.

## Alternatives

- Keep `ch start` running in the foreground: simpler, but it ties monitoring to an open terminal.
- Install a Windows service: robust, but too much setup and permission complexity for the first release.
- Use scheduled tasks or shell startup integration: useful later, but unnecessary for the basic command flow.

## Rationale

A detached worker keeps the user-facing CLI quick while allowing monitoring to continue in the background. Keeping worker process management inside `internal/monitor` makes the CLI layer smaller and keeps monitor lifecycle details together. The tradeoff is that the hidden `monitor` command is part of our internal process contract and must remain stable enough for `start` to launch it.
