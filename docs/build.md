# Build guide

## Local build

Use the provided script:

```bash
bash scripts/local-build.sh
```

This runs the test suite and produces a Windows portable executable in `dist/ch.exe`.

## Local debugging

Use the debug wrapper to rebuild the CLI and run a command:

```bash
./scripts/debug.sh start
./scripts/debug.sh status
```

All arguments are passed to the newly built `dist/ch.exe`. Before rebuilding,
the wrapper removes the previous binary. If that binary is locked by a running
clipboard monitor, the wrapper stops the monitor and retries the removal. It
exits with an error if the executable still cannot be removed.

Prefer this wrapper over `go run ./cmd/ch start` during development. The `start`
command launches a background copy of its own executable; on Windows, that keeps
Go's temporary executable locked and prevents `go run` from cleaning up its
temporary build directory. The background monitor also continues running the
last-built code until it is stopped and restarted.

## Docker build

```bash
docker build -t clip-history .
```

## Release build

Use the release script when you are ready to publish the current target version:

```bash
bash scripts/create-release.sh
```

The script validates the semver target in `VERSION`, injects that version into the release build, runs tests, builds `dist/ch.exe`, and creates the GitHub release with `gh`.
