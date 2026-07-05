# 0001: Use Go

Status: Accepted

Date: 2026-07-05

## Context

clip-history needs a small Windows-friendly executable with straightforward filesystem, process, and OS integration.

## Decision

Use Go for the application implementation.

## Alternatives

- Rust: strong safety and excellent binary distribution, but higher implementation complexity for this project.
- C#/.NET: strong Windows integration, but a heavier runtime and packaging story than we want right now.
- Python or Node.js: fast iteration, but less attractive for a portable single-binary CLI.

## Rationale

Go gives us simple cross-compilation, fast tests, a small standard-library-first toolchain, and easy single-binary distribution. The main tradeoff is that some Windows APIs require lower-level syscall work, but that is acceptable for the scope of this project.
