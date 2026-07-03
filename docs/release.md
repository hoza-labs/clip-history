# Release guide

## Strategy

clip-history uses semantic versioning. The checked-in `VERSION` file is the source of truth for the version we are currently working toward.

That target version is intentionally mutable until a GitHub release object is created. In practice:

- `VERSION` names the next intended release.
- `ch version` reads `VERSION` during normal development and reports the injected build version in release binaries.
- `CHANGELOG.md` keeps release notes under `Unreleased (target: x.y.z)` until the release is published.
- The version is only locked when `scripts/create-release.sh` creates the GitHub release and its `vX.Y.Z` tag.

## Prepare a release

1. Confirm `VERSION` contains the target semver, for example `0.0.1`.
2. Update `CHANGELOG.md` under `Unreleased (target: x.y.z)`.
3. Run `go run ./cmd/ch version` to verify the CLI reports the intended version.
4. Commit the release-ready state.

## Publish a release

Run:

```bash
bash scripts/create-release.sh
```

The script will:

- verify `gh` and `go` are installed
- verify `VERSION` is valid semver
- require a clean git working tree
- run `go test ./...`
- build `dist/ch.exe` with the release version injected
- create the GitHub release `vX.Y.Z` with generated notes and the Windows binary attached

## After a release

After GitHub has the release:

1. Move the released notes in `CHANGELOG.md` from `Unreleased` to a versioned section such as `## 0.0.1 - 2026-07-02`.
2. Bump `VERSION` to the next target version.
3. Add a fresh `Unreleased (target: next-version)` section for ongoing work.
4. Commit that post-release bump separately so the next cycle starts cleanly.
