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
- Track the current target release in [VERSION](VERSION) and follow [docs/release.md](docs/release.md) when publishing.
- Prefer independent unit tests with controlled dependencies; avoid accidental integration tests unless cross-package behavior is the thing being verified.
- Constructors should take all required dependencies for a minimally functional object. Optional settings may also be accepted by the constructor, and constructors with more than about five arguments should take a custom config/options struct instead.

## Local development

1. Install Go 1.22+.
2. Run tests: `go test ./...`
3. Build the binary: `go build ./cmd/ch`
4. Check the current target release: `go run ./cmd/ch version`
5. Start clipboard monitoring: `go run ./cmd/ch start`

## Structure

- cmd/ch: CLI entrypoint.
- internal/cli: command tree and subcommands.
- internal/config: home-directory config handling.
- internal/logging: file-based logging.
- internal/monitor: background monitor state and Windows clipboard listener.
- internal/safety: sensitivity filtering.
- internal/version: version resolution for development and release builds.
- docs: user documentation for GitHub Pages.
