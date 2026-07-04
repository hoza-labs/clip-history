# Conventions

Shared coding and design rules for this repository live here. Keep this file as the canonical source for conventions that apply to both humans and agents.

## Documentation layout

- Keep shared coding and design rules in `CONVENTIONS.md`.
- Reserve `AGENTS.md` for truly agent-specific instructions.
- Use `README.md` for project overview, development workflow, and links to the canonical conventions.

## Project rules

- Keep the repository simple, modular, and test-driven.
- Prefer small files, small functions, and public-interface tests.
- Prefer independent unit tests with controlled dependencies; avoid integration tests unless cross-package or cross-module behavior is the explicit target.
- Prefer standard library solutions over third-party dependencies unless there is strong justification.
- Constructors should take all required dependencies for a minimally functional object.
- Constructors may also accept optional settings.
- Constructors with more than about five arguments should take a single custom config/options object instead.
- Favor avoiding duplication by storing a normalized copy of constructor config objects rather than repeating the same fields on constructed types.
- Package methods and functions should prefer accepting interfaces and returning structs.
- Exported structs should generally favor keeping fields private, exposing behavior through methods rather than public mutable fields.
