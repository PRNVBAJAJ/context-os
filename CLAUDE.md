# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Context OS is a **local-first runtime** that provides persistent project state, durable workflow execution, and shared memory across AI coding assistants (Claude Code, Codex CLI, Gemini CLI, OpenCode, etc.). It is not another coding assistant — it is the operating system beneath them. The central thesis: project intelligence belongs to the project, not the assistant.

The repository is currently in the **design/architecture phase**. All implementation files (Go source, `go.mod`, `Makefile`) are yet to be created. The `docs/architecture/` directory contains the authoritative 27-chapter design specification.

## Planned Technology Stack

| Layer         | Technology                          |
|---------------|-------------------------------------|
| Language      | Go                                  |
| CLI Framework | Cobra                               |
| Config        | Viper                               |
| TUI           | Bubble Tea + Lip Gloss              |
| Database      | SQLite (embedded, ACID)             |
| Artifacts     | Markdown + JSON + JSONL             |
| Logging       | Go `slog`                           |
| Markdown      | Goldmark                            |
| Releases      | GoReleaser                          |
| CI            | GitHub Actions                      |

## Planned Commands (once implemented)

```bash
# Build
go build ./cmd/context/...

# Run all tests
go test ./...

# Run a single test
go test ./internal/<package>/... -run TestName

# Lint
golangci-lint run

# Build release artifacts
goreleaser build --snapshot
```

## Planned Repository Layout

```
cmd/context/main.go          # Entrypoint — wires deps, calls cli.Execute()
internal/
  application/               # Use cases: StartWorkflow, ResumeWorkflow, CreateCheckpoint
  runtime/                   # RuntimeManager, RuntimeState — owns execution lifecycle
  workflow/                  # Core engine: Workflow, Step, WorkflowState, WorkflowManager
  contextbuilder/            # Assembles ExecutionContext from workflow+memory+artifacts
  memory/                    # Durable project knowledge (architecture, decisions, conventions)
  artifact/                  # Generated outputs (designs, reviews, plans, benchmarks)
  checkpoint/                # Create / restore / archive execution recovery points
  provider/                  # Provider interface definitions and ProviderRegistry
  adapter/                   # Concrete CLI adapters: Claude, Codex, Gemini, OpenCode, Shell
  session/                   # Session lifecycle: resume, pause, interrupt, recover
  storage/                   # Persistence abstraction over SQLite / Markdown / JSON
  event/                     # Immutable audit events: WorkflowStarted, ArtifactCreated, etc.
  config/                    # Config loading: global → project → env vars → CLI flags
  cli/                       # Cobra command definitions (no business logic)
  tui/                       # Bubble Tea views, models, update loop
  plugin/                    # Plugin loading, manifest validation, capability registration
  shared/                    # Errors, constants, IDs, validation — kept intentionally small
pkg/                         # Stable public SDKs (Provider SDK, Plugin SDK)
api/                         # Reserved for future REST / gRPC / MCP (empty in v0.1)
configs/                     # Default configs, provider templates, workflow templates
test/                        # Large-scale integration and recovery tests
```

## Architecture Rules

**Layer order (no bypassing):** Presentation → Application → Domain → Infrastructure → Storage

**Dependency direction:** `cmd` → `cli` → `application` → `workflow` → `storage/provider` → `adapter` → shell. Only downward dependencies are allowed.

**Package ownership is exclusive** — no two packages own the same domain:

| Package        | Owns                        |
|----------------|-----------------------------|
| `workflow`     | Workflow lifecycle           |
| `runtime`      | Current execution state      |
| `memory`       | Project knowledge            |
| `artifact`     | Generated outputs            |
| `checkpoint`   | Recovery / restore           |
| `contextbuilder` | Prompt/context assembly    |
| `adapter`      | Provider-specific execution  |
| `storage`      | Persistence only             |

## Core Design Decisions

- **Runtime, not framework** — planning stays with the external assistant; Context OS executes.
- **Adapters only** — the runtime never invokes AI providers directly, always through an adapter.
- **Workflows over conversations** — structured workflow state is the source of truth, never conversation history.
- **Failure isolation** — a provider crash must not corrupt workflow, checkpoint, or memory state.
- **Human-readable storage** — SQLite for structured metadata/indexes; Markdown for artifacts and memory; JSON/JSONL for interchange. Developers must be able to inspect project state without Context OS installed.
- **No DI frameworks** — use Go interfaces for dependency inversion; keep it idiomatic.

## Runtime Directory (`.context/`)

`context init` creates `.context/` inside the user's project with:

```
runtime.db       # SQLite — workflows, sessions, events, metadata
project.yaml     # Project configuration
memory/          # Markdown knowledge files
artifacts/       # Generated outputs
sessions/        # Session state
workflows/       # Workflow definitions
checkpoints/     # Recovery snapshots
logs/
cache/
providers/
```

## v0.1 MVP Scope

In scope: local-only, CLI providers, sequential workflows, checkpoint-based recovery, durable memory (Markdown + metadata), basic TUI dashboard.

Out of scope: cloud sync, team collaboration, vector/semantic search, embeddings, REST/MCP adapters, marketplace, enterprise security, autonomous execution.

**Performance targets:** runtime startup < 200 ms, workflow restore < 500 ms, context assembly < 250 ms, checkpoint creation < 100 ms, provider dispatch overhead < 50 ms.

## Key Architecture Documents

The `docs/architecture/` directory is the canonical specification. Most relevant chapters:

- `01-executive-summary.md` — problem statement and vision
- `05-highlevel-architecture.md` — system context and component overview
- `06-layered-architecture.md` — layer boundaries and dependency rules
- `07-technology-stack.md` — technology decisions and rationale
- `08-repository-structure.md` — package layout and ownership rules
- `13-execution-lifecycle.md` — command lifecycle from parse to result
- `17-adapter-framework.md` — how providers are integrated
- `18-workflow-engine.md` — workflow state machine design
- `23-mvp-scope.md` — what is and is not in v0.1
