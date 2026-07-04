# AGENTS

- Keep this repository simple, modular, and test-driven.
- Prefer small files, small functions, and public-interface tests.
- Prefer independent unit tests with controlled dependencies; avoid integration tests unless cross-module behavior is the explicit target.
- Constructors should take all required dependencies for a minimally functional object. Constructors may also accept optional settings, and constructors with more than about five arguments should take a single custom config/options object instead.
- Favor avoiding duplication by storing a normalized copy of constructor config objects rather than repeating the same fields on constructed types.
- Favor exporting an interface and keeping concrete implementation details private when that preserves flexibility for future changes.
- Update README, PLAN.md, and CHANGELOG.md whenever the project direction changes.
- Prefer standard library solutions over third-party dependencies unless there is strong justification.
- Treat `VERSION` as the current target semver release and follow `docs/release.md` when preparing or publishing releases.
