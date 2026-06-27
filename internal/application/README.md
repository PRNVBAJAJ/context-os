# internal/application

Application-layer use cases for Context OS.

This package coordinates domain objects and infrastructure — it contains no business rules. Each function in this package corresponds to a user-facing operation and sequences the domain + storage calls needed to complete it.

## Use Cases (current)

- **`InitializeProject`** — implements `context init`: validates inputs, creates the `.context/` layout, writes `project.yaml`, bootstraps SQLite, and persists the project record.

## Rules

- No Cobra imports.
- No business rules (validation lives in domain packages).
- No global state.
- All functions accept `context.Context` as the first argument.
- Tests use real temp directories and real SQLite — no mocks.
