# internal/event

Immutable audit event definitions for Context OS.

Events record what happened in the runtime. They are append-only — never modified or deleted after creation. Every significant state change (project initialized, workflow started, checkpoint created, etc.) produces an event.

## Contents

- **`event.go`** — `Event` struct, `Type` string type, constants, `New()` constructor.

## Rules

- Imports only `internal/shared`.
- Events are value objects: construct with `New()`, store via `storage.EventStore`.
- `WorkflowID` is empty (`shared.EmptyID`) for project-level events.
- `Payload` must be valid JSON; use `"{}"` when no additional context is needed.

## Adding new event types

Add a `const` in `event.go`. The payload schema is defined by convention (document it in the constant's comment). No code generation required.
