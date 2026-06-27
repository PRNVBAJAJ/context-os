# Contributing to Context OS

## Before you start

- Open an issue before submitting large changes so the approach can be agreed on first.
- All code must compile (`go build ./...`) and pass tests (`go test ./...`) before review.
- Follow the layered architecture described in `CLAUDE.md` — no upward imports between layers.

## Development setup

```bash
git clone https://github.com/context-os/context-os
cd context-os
make build
make test
```

Requires Go 1.25+. No CGO, no external runtime dependencies.

## Standards

**No scaffolding.** Every function that is committed must be fully implemented. No TODO bodies, no placeholder returns.

**No mocks.** Tests use real SQLite databases in `t.TempDir()`. Integration tests are fast enough that mocks add no value.

**Table-driven tests** for validation logic. Use `t.Run` subtests.

**Layered architecture** (enforced by import rules):

```
cli → application → domain (project, workflow, memory, checkpoint, event) → storage
```

Domain packages import only `internal/shared`. Storage imports domain. Application imports domain + storage. CLI imports application + project (for `Dir()`). TUI imports domain only.

**Error handling.** Return `*shared.Error` with an appropriate `Code` from domain and application code. Wrap external errors with `shared.Wrap`. Never return raw `fmt.Errorf` from domain or application layers.

**No global state.** Pass dependencies explicitly. No package-level `var` that mutates after init.

## Adding a new command

1. Domain: add to the appropriate `internal/<domain>` package.
2. Storage: add a migration (append-only, never modify existing migrations) and a store method.
3. Application: add a use case function in `internal/application/`.
4. CLI: add a `new<Name>Command()` in `internal/cli/`, register it in `root.go`.
5. Tests at every layer.

## Commit style

```
<type>: <short imperative description>

<optional body explaining why, not what>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.

## Pull requests

- One logical change per PR.
- Include tests for any new behaviour.
- Update `CLAUDE.md` if you change architecture rules, planned commands, or the tech stack.
