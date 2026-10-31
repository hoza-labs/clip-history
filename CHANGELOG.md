# Changelog

## Unreleased (target: 0.0.1)

- Removed the unused sensitive-content storage setting from generated configuration and `ch config` output.
- Changed `ch config` to display all effective configuration parameters as TOML, including defaults and normalized values, and added a convention requiring output and tests to stay current with configuration changes.
- Added progressive-disclosure guidance for organizing source files around their public behavior.
- Added detailed lifecycle decision logging for monitor start, stop, worker shutdown, and log retention cleanup.
- Improved logger failure handling so retention cleanup failures are warnings, unwritable entries fall back to stderr, and negative retention settings are rejected.
- Made the monitor controller logger an explicit required dependency.
- Made monitor state coordination concurrency-safe with serialized access, atomic file replacement, bounded Windows sharing retries, and joined heartbeat shutdown.

- Scaffolded the initial Go CLI project structure.
- Added config persistence under the user's home directory.
- Added file-based logging and sensitivity filtering modules.
- Added date-stamped log files with configurable retention cleanup.
- Added configurable log time zones with a default of local time.
- Added automatic logger rotation at midnight with retention cleanup on rotation.
- Added millisecond log timestamps and command titles to log entries.
- Added shared CLI output handling that writes command output to stdout and logs it.
- Added initial BDD-style unit tests for the public module behavior.
- Added `ch version` and a single source of truth for the target semver release.
- Added a documented GitHub release workflow driven by `scripts/create-release.sh`.
- Added `ch start`, `ch stop`, and `ch status` for background clipboard monitoring.
- Added a Windows clipboard listener process that logs clipboard copy events and maintains monitor state.
- Tightened constructor guidance and refactored monitor service creation to require all minimal dependencies at construction time.
- Refactored monitor control behind an exported interface with a private implementation that stores normalized constructor config.
- Consolidated shared repository rules into `CONVENTIONS.md` and reserved `AGENTS.md` for agent-specific instructions.
- Updated the Go API design convention to prefer accepting interfaces and returning structs, and refactored monitor service accordingly.
- Renamed monitor lifecycle coordination to `Controller` and the long-running child process implementation to `Worker`.
- Adopted Architecture Decision Records and documented the initial language, CLI, background worker, and worker coordination decisions.
- Added a Go documentation convention and brief comments for packages, functions, methods, and package-level declarations.
- Switched user configuration from JSON to TOML with mixed-case keys using the Pelletier go-toml v2 library.
