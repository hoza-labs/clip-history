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
- Follow the shared project rules in [CONVENTIONS.md](CONVENTIONS.md).
- Review architecture decisions in [docs/adr](docs/adr/index.md).
- Keep agent-facing instructions current in [AGENTS.md](AGENTS.md) and [SKILLS.md](SKILLS.md).
- Track the current target release in [VERSION](VERSION) and follow [docs/release.md](docs/release.md) when publishing.

## Local development

Use `ch config` to display every effective configuration parameter as TOML, including defaults and normalized values.

1. Install Go 1.22+.
2. Run tests: `go test ./...`
3. Build the binary: `go build ./cmd/ch`
4. Check the current target release: `go run ./cmd/ch version`
5. Build and start clipboard monitoring for debugging: `./scripts/debug.sh start`

Use `scripts/debug.sh` when exercising the CLI during development:

```bash
./scripts/debug.sh <command> [arguments...]
```

The script runs the local build, creates `dist/ch.exe`, and forwards all of its
arguments to that executable. If an earlier `dist/ch.exe` monitor is still
running, the script stops it before rebuilding. This avoids the temporary-file
cleanup errors caused by `go run ./cmd/ch start` on Windows. A background monitor
continues running the code from its last build, so rerun the debug script after
source changes when you want to test the updated monitor.

## Structure

- cmd/ch: CLI entrypoint.
- internal/cli: command tree and subcommands.
- internal/config: home-directory config handling, including log retention and time zone settings.
- internal/logging: date-stamped file-based logging with retention cleanup and configurable timestamps.
- internal/monitor: monitoring controller plus Windows clipboard worker implementation.
- internal/safety: sensitivity filtering.
- internal/version: version resolution for development and release builds.
- docs: user documentation for GitHub Pages.
