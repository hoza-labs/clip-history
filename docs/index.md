# clip-history documentation

This site will host build, install, and usage guidance for clip-history.

## Getting started
- Build locally with the script in [scripts/local-build.sh](../scripts/local-build.sh).
- Run the CLI with `go run ./cmd/ch --help`.

## Installation
- Download the portable binary from release artifacts.
- Place it in a directory already on your PATH.

## Usage
- `ch config` shows configuration details.
- `ch start` starts background clipboard monitoring.
- `ch stop` stops background clipboard monitoring.
- `ch status` shows the current history directory and whether monitoring is active.
- `ch version` shows the current target release version.

## Releases
- Track the next release in [VERSION](../VERSION).
- Publish with [scripts/create-release.sh](../scripts/create-release.sh).
