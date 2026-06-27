# internal/cli

Cobra command definitions for the `context` binary.

This package is the **Presentation Layer** boundary. It owns command parsing and flag definitions, then delegates to `internal/application` use cases (added in future milestones). It never contains business logic.

## Dependency rule

`internal/cli` may import:
- `internal/shared` (primitives)
- `internal/application` (use cases, added later)
- `github.com/spf13/cobra`

It must never import domain or infrastructure packages directly.

## Command tree (current)

```
context
  --version    Print the runtime version and exit
  --help       Print usage and exit
```

Subcommands (`init`, `workflow`, `checkpoint`, etc.) are registered here as each milestone is implemented.
