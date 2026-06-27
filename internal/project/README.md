# internal/project

Domain package for the Context OS project entity.

Owns three responsibilities:

- **`project.go`** — `Project` struct (aggregate root) and `New()` constructor with validation. The only package allowed to create a `Project`.
- **`layout.go`** — `.context/` directory constants, `CreateLayout()`, `IsInitialized()`. Defines the canonical subdirectory structure that every other runtime service depends on.
- **`yaml.go`** — `Save()` / `Load()` for `project.yaml`. Note: `root_path` is never written to disk; it is derived from the file's location, keeping the file portable.

## Rules

- No imports from `internal/storage`, `internal/application`, or `internal/cli`.
- Only imports `internal/shared` (for `ID`, `Version`, error types).
- All business validation lives here — not in the CLI or application layer.
