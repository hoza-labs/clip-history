# Conventions

Shared coding and design rules for this repository live here. Keep this file as the canonical source for conventions that apply to both humans and agents.

## Documentation layout

- Keep shared coding and design rules in `CONVENTIONS.md`.
- Reserve `AGENTS.md` for truly agent-specific instructions.
- Use `README.md` for project overview, development workflow, and links to the canonical conventions.
- Track significant architecture decisions as brief ADRs in `docs/adr/`.
- ADRs should include status, date, context, decision, top alternatives, and brief rationale with pros and cons.

## Project rules

- Keep the repository simple, modular, and test-driven.
- Document every Go package, function, method, type, and package-level var or const with a brief comment; mention inputs and return values when they are not obvious from the name.
- Package comments should summarize the package protocol, including required call sequencing, important conditions, and private/specialized entrypoints such as hidden CLI commands or worker methods.
- Constructor comments should state required dependencies and lifecycle protocols; if a returned value owns resources, tell callers to call `Close`, usually with `defer value.Close()` immediately after checking the error.
- Prefer small files, small functions, and public-interface tests.
- Organize source files for progressive disclosure: present primary types, constructors, and public behavior before internal orchestration and low-level helpers. Keep supporting declarations close to the behavior they explain, and deviate when dependency-first ordering or grouping materially improves clarity.
- Prefer independent unit tests with controlled dependencies; avoid integration tests unless cross-package or cross-module behavior is the explicit target.
- Prefer functions to log significant decisions, branches, and return paths so runtime behavior can be reconstructed during debugging.
- Prefer standard library solutions over third-party dependencies unless there is strong justification.
- Constructors should take all required dependencies for a minimally functional object.
- Constructors may also accept optional settings.
- Constructors with more than about five arguments should take a single custom config/options object instead.
- Favor avoiding duplication by storing a normalized copy of constructor config objects rather than repeating the same fields on constructed types.
- Package methods and functions should prefer accepting interfaces and returning structs.
- Exported structs should generally favor keeping fields private, exposing behavior through methods rather than public mutable fields.
- Constructors should generally return values (plain structs) rather than pointers to structs, but there are exceptions for very large structs and most structs that themselves contain pointers.
- Constructors should return pointers to structs that own mutable state, mutexes, file handles, connections, caches, or other resources that should not be copied.
- Structs should be passed by value whenever they are intended to be immutable.
