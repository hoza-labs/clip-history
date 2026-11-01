# Plan

## Phase 1

- [x] Scaffold a Go CLI project structure.
- [x] Add config and logging modules with tests.
  Current progress: logs are date-stamped, rotate at midnight, cleaned up according to config retention, and written in the configured time zone.
- [x] Add a richer CLI command tree with subcommands and help.
- [x] Display every effective configuration parameter through `ch config` and require output coverage as configuration evolves.
- [x] Establish semver-based version tracking and a GitHub release workflow.
- [x] Adopt Architecture Decision Records for major project decisions.
- [x] Assess dependency supply-chain risks and document dependency retention decisions and a seven-day update cooldown in [SECURITY-REVIEW.md](SECURITY-REVIEW.md).
- [ ] Update and pin build inputs, complete vulnerability scans, and add declarative dependency update proposals with independent release-age and vulnerability checks; retain maintainer review for trust decisions and exceptions.
- [ ] Add clipboard capture and history storage workflows.
  Current progress: background clipboard monitoring can be started and stopped, and clipboard copy events are logged.
- [ ] Add Docker-based local build workflow and lightweight CI.
- [ ] Produce install, build, and usage documentation.
