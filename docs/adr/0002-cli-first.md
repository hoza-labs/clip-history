# 0002: Build a CLI-first application

Status: Accepted

Date: 2026-07-05

## Context

The first release needs reliable clipboard monitoring and local history behavior before any visual interface.

## Decision

Make `ch` a CLI-first application.

## Alternatives

- Desktop GUI first: more discoverable, but larger scope and more design/runtime complexity.
- Background-only daemon: simpler surface, but harder for users to inspect, configure, and troubleshoot.
- Web UI first: flexible, but not a natural fit for a local Windows clipboard utility.

## Rationale

A CLI keeps the initial product small, scriptable, testable, and easy to distribute. It also gives users explicit commands for `start`, `stop`, `status`, `config`, and `version`. The tradeoff is that some users may eventually want a GUI, but the CLI gives us a strong foundation first.
