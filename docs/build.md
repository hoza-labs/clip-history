# Build guide

## Local build

Use the provided script:

```bash
bash scripts/local-build.sh
```

This runs the test suite and produces a Windows portable executable in `dist/ch.exe`.

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
