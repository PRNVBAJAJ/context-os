# internal/storage

Persistence abstraction for Context OS.

Exposes two interfaces (`Storage`, `ProjectStore`) and a single constructor (`Open`). All callers work against the interfaces — the SQLite implementation is an internal detail.

## Interfaces

- **`Storage`** — top-level handle; gateway to all sub-stores; must be `Close()`d by caller.
- **`ProjectStore`** — CRUD for project metadata (the `project` table).

## Adding a new migration

Append a new entry to the `migrations` slice in `migrations.go`. Increment `version` by 1. Never modify existing migration entries — they are immutable once applied to any database.

## Rules

- Never import `internal/application` or `internal/cli`.
- All time values stored and returned in UTC, RFC3339 format.
- All errors returned as `*shared.Error` with appropriate `Code`.
- SQLite connection pool limited to 1 writer (WAL mode is a future enhancement).
