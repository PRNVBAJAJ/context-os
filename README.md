# Context OS

**Provider-agnostic runtime for AI-assisted engineering.**

Context OS is a local-first runtime that sits beneath your AI coding assistant — not beside it. It owns your project's engineering intelligence: structured workflows, durable memory, recovery checkpoints, and an immutable audit log. When you switch from Claude to Codex to Gemini and back, your context travels with the project, not with any provider.

```
Git       → source code history
Context OS → engineering intelligence
AI Provider → execution engine
```

## Install

### Homebrew (macOS / Linux)

```bash
brew install context-os/tap/context
```

### Pre-built binaries

Download the latest release for your platform from the [Releases](https://github.com/context-os/context-os/releases) page. Binaries are statically linked — no runtime dependencies.

### Build from source

Requires Go 1.25+.

```bash
git clone https://github.com/context-os/context-os
cd context-os
make build          # produces bin/context
```

## Quick start

```bash
# Initialise a project
cd my-project
context init --name "my-project" --language go

# Start a workflow
context workflow start "implement auth"

# Add durable memory
context memory add jwt-strategy "RS256, keys rotate every 90 days"

# Create a recovery checkpoint
context checkpoint create --workflow <id> --note "skeleton done, need refresh tokens"

# Open the dashboard
context tui

# Check runtime health
context doctor
```

## Commands

| Command | Description |
|---------|-------------|
| `context init` | Initialise a Context OS project in the current directory |
| `context status` | Show current project metadata |
| `context doctor` | Health check: runtime, database, recent events |
| `context tui` | Interactive dashboard (workflows, memories, events) |
| **Memory** | |
| `context memory add <key> <content>` | Add a named knowledge entry |
| `context memory list` | List all memory entries |
| **Workflows** | |
| `context workflow start <name>` | Create and start a new workflow |
| `context workflow list` | List all workflows |
| `context workflow complete <id>` | Mark a workflow completed |
| `context workflow fail <id>` | Mark a workflow failed |
| `context workflow pause <id>` | Pause a running workflow |
| `context workflow resume <id>` | Resume a paused workflow |
| **Checkpoints** | |
| `context checkpoint create` | Snapshot current state for recovery |
| `context checkpoint list` | List all checkpoints |

Workflow commands accept an ID prefix (first 8 characters) instead of the full UUID.

## Runtime directory

`context init` creates `.context/` inside your project:

```
.context/
  runtime.db       # SQLite — workflows, sessions, events, metadata
  project.yaml     # Project configuration
  memory/          # Markdown knowledge files (human-readable)
  artifacts/       # Generated outputs
  sessions/        # Session state
  workflows/       # Workflow definitions
  checkpoints/     # Recovery snapshots
  logs/
```

All persistent state is inspectable without Context OS installed. `runtime.db` is a standard SQLite file; `.context/memory/*.md` files are plain Markdown.

## Architecture

```
Presentation  internal/cli     — Cobra commands
Application   internal/application — Use cases
Domain        internal/project, workflow, memory, checkpoint, event
Storage       internal/storage — SQLite via modernc.org/sqlite (no CGO)
TUI           internal/tui     — Bubble Tea dashboard
```

Strict layered dependency: each layer only imports downward. The domain never imports storage or infrastructure. See `docs/architecture/` for the full 27-chapter specification.

## Development

```bash
make build    # build bin/context
make test     # go test ./...
make lint     # golangci-lint run
make clean    # remove bin/
```

Run a single test:

```bash
go test ./internal/<package>/... -run TestName
```

Integration tests use real SQLite databases in `t.TempDir()` — no mocks, no network.

## Performance targets (v0.1)

| Operation | Target |
|-----------|--------|
| Runtime startup | < 200 ms |
| Workflow restore | < 500 ms |
| Context assembly | < 250 ms |
| Checkpoint creation | < 100 ms |
| Provider dispatch overhead | < 50 ms |

## License

MIT — see [LICENSE](LICENSE).
