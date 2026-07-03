# AGENTS

- Keep this repository simple, modular, and test-driven.
- Prefer small files, small functions, and public-interface tests.
- Prefer independent unit tests with controlled dependencies; avoid integration tests unless cross-module behavior is the explicit target.
- Update README, PLAN.md, and CHANGELOG.md whenever the project direction changes.
- Prefer standard library solutions over third-party dependencies unless there is strong justification.
- Treat `VERSION` as the current target semver release and follow `docs/release.md` when preparing or publishing releases.
