# internal/shared

Foundation primitives used by every other package in Context OS.

This package is intentionally small. It provides only three things:

- **`Error` / `Code`** — the standard domain error type. All packages return `*shared.Error` at package boundaries so callers can inspect the error class (`Code`) without importing unrelated packages.
- **`ID`** — a typed UUID v4 string used as the canonical identifier for all domain objects (workflows, sessions, checkpoints, artifacts, etc.).
- **`Version`** — the current runtime version string, used by the CLI and surfaced via `context --version`.

## Rules

- No imports from any other `internal/` package.
- No business logic.
- No I/O.
- No global state.
