# Changelog

## Unreleased (target: 0.0.1)

- Scaffolded the initial Go CLI project structure.
- Added config persistence under the user's home directory.
- Added file-based logging and sensitivity filtering modules.
- Added initial BDD-style unit tests for the public module behavior.
- Added `ch version` and a single source of truth for the target semver release.
- Added a documented GitHub release workflow driven by `scripts/create-release.sh`.
- Added `ch start`, `ch stop`, and `ch status` for background clipboard monitoring.
- Added a Windows clipboard listener process that logs clipboard copy events and maintains monitor state.
- Tightened constructor guidance and refactored monitor service creation to require all minimal dependencies at construction time.
- Refactored monitor control behind an exported interface with a private implementation that stores normalized constructor config.
- Consolidated shared repository rules into `CONVENTIONS.md` and reserved `AGENTS.md` for agent-specific instructions.
