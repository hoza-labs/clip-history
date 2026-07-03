# clip-history

clip-history (ch) is a command-line clipboard manager for Windows that preserves clipboard items on disk for later review and reuse while helping prevent accidental storage of sensitive data.

## Project goals

- Build a portable CLI-first application for Windows 11.
- Keep the implementation modular and test-driven.
- Favor simple, dependency-light code and local Docker-based builds.
- Generate docs in the docs folder for GitHub Pages.

## Repository guidance

- Follow the plan in [PLAN.md](PLAN.md).
- Track shipped work in [CHANGELOG.md](CHANGELOG.md).
- Keep agent-facing instructions current in [AGENTS.md](AGENTS.md) and [SKILLS.md](SKILLS.md).

## Local development

1. Install Go 1.22+.
2. Run tests: `go test ./...`
3. Build the binary: `go build ./cmd/ch`

## Structure

- cmd/ch: CLI entrypoint.
- internal/cli: command tree and subcommands.
- internal/config: home-directory config handling.
- internal/logging: file-based logging.
- internal/safety: sensitivity filtering.
- docs: user documentation for GitHub Pages.
