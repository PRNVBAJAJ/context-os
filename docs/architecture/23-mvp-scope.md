# Chapter 23 — MVP Scope (Version 0.1)

---

# Chapter 23 — MVP Scope (Version 0.1)

## 23.1 Overview

One of the biggest reasons engineering projects fail is **trying to build Version 2 first**.

Throughout this design document, we have described a long-term vision for Context OS:

* Multi-agent workflows
* Knowledge graphs
* Team collaboration
* Cloud synchronization
* MCP
* Semantic search
* Plugins
* Enterprise capabilities

None of these belong in Version 0.1.

The purpose of Version 0.1 is much simpler:

> **Prove that project intelligence can exist independently of AI conversations.**

If Version 0.1 achieves that, the remaining architecture becomes evolutionary rather than revolutionary.

---

# 23.2 Product Vision for v0.1

Version 0.1 should solve one problem exceptionally well.

> **Provide persistent project context and resumable workflows for CLI-based AI coding assistants.**

Supported providers include:

* Claude Code
* Codex CLI
* Gemini CLI
* OpenCode
* Any future CLI compatible with the Adapter interface

Version 0.1 is intentionally **CLI-first**.

---

# 23.3 Goals

Version 0.1 must achieve the following.

✓ Local First

✓ Provider Agnostic

✓ Workflow Driven

✓ Context Reconstruction

✓ Durable Memory

✓ Checkpoint Recovery

✓ Human Readable Storage

✓ Open Source

---

# 23.4 Non-Goals

The following features are **explicitly excluded** from Version 0.1.

✗ Cloud synchronization

✗ Team collaboration

✗ Knowledge Graph

✗ Vector Search

✗ Embeddings

✗ API providers

✗ Distributed execution

✗ Marketplace

✗ Enterprise security

✗ Organization memory

✗ Autonomous engineering

If a feature is not required to prove the core concept, it does not belong in v0.1.

---

# 23.5 User Persona

Primary audience:

```text
Individual software engineers
```

Examples:

* Staff Engineers
* Open Source Contributors
* Startup Founders
* AI-first Developers
* Freelancers
* Platform Engineers

Not optimized for enterprise teams.

---

# 23.6 Supported Platforms

Version 0.1 officially supports:

| Platform       | Support      |
| -------------- | ------------ |
| macOS          | ✓            |
| Linux          | ✓            |
| Windows (WSL)  | ✓            |
| Native Windows | Experimental |

---

# 23.7 Supported Providers

Version 0.1 supports CLI-based providers only.

| Provider          | Support |
| ----------------- | ------- |
| Claude Code       | ✓       |
| OpenCode          | ✓       |
| Codex CLI         | ✓       |
| Gemini CLI        | ✓       |
| Generic Shell CLI | ✓       |

Adapters are responsible for provider-specific behavior.

---

# 23.8 Core Commands

The CLI should expose a minimal command surface.

```bash
context init

context status

context workflow start

context workflow list

context workflow resume

context workflow pause

context memory list

context memory show

context checkpoint list

context checkpoint restore

context provider list

context doctor
```

This is sufficient for Version 0.1.

---

# 23.9 Runtime Directory

Running

```bash
context init
```

creates

```text
.context/

runtime.db

project.yaml

memory/

artifacts/

sessions/

workflows/

checkpoints/

logs/

cache/

providers/
```

No additional directories should be required.

---

# 23.10 Core Runtime Components

Version 0.1 includes exactly the following services.

| Component          | Included |
| ------------------ | -------- |
| Runtime            | ✓        |
| Workflow Engine    | ✓        |
| Context Builder    | ✓        |
| Session Manager    | ✓        |
| Checkpoint Manager | ✓        |
| Memory Manager     | ✓        |
| Artifact Manager   | ✓        |
| Adapter Manager    | ✓        |
| Provider Registry  | ✓        |
| Storage Manager    | ✓        |

No scheduler.

No distributed runtime.

---

# 23.11 Storage

Version 0.1 storage model:

| Technology | Included |
| ---------- | -------- |
| SQLite     | ✓        |
| Markdown   | ✓        |
| JSON       | ✓        |
| JSONL      | ✓        |
| Cache      | ✓        |

No vector database.

No cloud database.

---

# 23.12 Context Builder

Version 0.1 Context Builder assembles context using:

✓ Workflow

✓ Session

✓ Checkpoint

✓ Memory

✓ Artifacts

✓ Repository Summary

No embeddings.

No semantic retrieval.

No LLM-generated memory synthesis.

---

# 23.13 Memory

Memory supports:

✓ Markdown

✓ Metadata

✓ Tags

✓ Retrieval

✓ Archiving

No automatic summarization.

No vector indexing.

---

# 23.14 Workflow Engine

Supports:

✓ Sequential workflows

✓ DAG representation

✓ Checkpoints

✓ Retry

✓ Resume

Future parallel execution is intentionally disabled.

---

# 23.15 Checkpoints

Version 0.1 supports

```bash
context checkpoint create

context checkpoint list

context checkpoint restore
```

Checkpoints capture:

* Workflow state
* Session state
* Current step
* Context references

Not provider conversation history.

---

# 23.16 Adapter Framework

Supported adapters:

```text
Claude Adapter

Codex Adapter

Gemini Adapter

OpenCode Adapter

Shell Adapter
```

No REST adapters.

No MCP adapters.

---

# 23.17 TUI

Version 0.1 includes a lightweight terminal UI.

Views:

```text
Dashboard

Workflow

Memory

Artifacts

Logs

Providers
```

The TUI complements the CLI.

The CLI remains the primary interface.

---

# 23.18 Events

Version 0.1 records immutable runtime events.

Examples:

* WorkflowStarted
* WorkflowCompleted
* CheckpointCreated
* ArtifactGenerated
* ProviderExecuted

These power debugging and future analytics.

---

# 23.19 Security

Version 0.1 includes:

✓ Local storage

✓ Permission validation

✓ Runtime ownership of `.context`

✓ Provider isolation

✓ Audit events

Future encryption and RBAC are deferred.

---

# 23.20 Plugins

Plugin support is intentionally minimal.

Included:

* Discovery
* Manifest validation
* Capability registration

Excluded:

* Marketplace
* Hot reload
* Dependency resolution
* Remote plugins

---

# 23.21 Error Recovery

The runtime must recover from:

* Provider crashes
* Context limit exhaustion
* Interrupted execution
* Runtime restart

Recovery is checkpoint-based.

---

# 23.22 Performance Targets

The following targets guide implementation.

| Metric                     | Target   |
| -------------------------- | -------- |
| Runtime startup            | < 200 ms |
| Workflow restore           | < 500 ms |
| Context assembly           | < 250 ms |
| Checkpoint creation        | < 100 ms |
| SQLite initialization      | < 100 ms |
| Provider dispatch overhead | < 50 ms  |

These are engineering goals rather than hard guarantees.

---

# 23.23 Out-of-Scope Decisions

The following architectural questions are intentionally postponed.

* Which embedding model?
* Which vector database?
* Team synchronization protocol
* Remote execution
* Marketplace design
* Knowledge graph schema
* Cloud storage provider

Premature decisions would unnecessarily constrain future evolution.

---

# 23.24 Success Criteria

Version 0.1 is considered successful if a developer can:

1. Initialize a project.
2. Configure one or more CLI providers.
3. Start a workflow.
4. Interrupt execution.
5. Resume later.
6. Switch providers.
7. Continue without losing project context.
8. Inspect memory, artifacts, checkpoints, and workflow state.
9. Complete the workflow without relying on conversation history.

If these scenarios work reliably, the central hypothesis of Context OS has been validated.

---

# 23.25 Acceptance Test

A canonical end-to-end scenario.

```text
Developer

↓

context init

↓

Configure Providers

↓

Start OAuth Workflow

↓

Claude plans implementation

↓

Checkpoint Created

↓

Switch to Codex

↓

Codex implements feature

↓

Checkpoint Created

↓

Claude reviews implementation

↓

Workflow Completed

↓

Artifacts Stored

↓

Memory Updated

↓

Project Closed
```

At no point should provider conversation history be required.

---

# 23.26 Release Deliverables

Version 0.1 ships with:

* Context OS CLI
* Bubble Tea TUI
* SQLite storage
* Workflow engine
* Context builder
* Checkpoint manager
* Memory manager
* Artifact manager
* Provider adapters (CLI)
* Documentation
* Examples
* Plugin SDK (minimal)

Everything else is deferred.

---

# 23.27 Design Decisions

## Decision 1 — Solve One Problem Well

Version 0.1 focuses exclusively on durable project context for CLI-based AI assistants.

---

## Decision 2 — Local First

All functionality must work completely offline except for the provider itself.

---

## Decision 3 — Stable Core

Future capabilities should extend the runtime rather than requiring architectural rewrites.

---

## Decision 4 — Human Readable by Default

Project knowledge remains inspectable and editable without specialized tooling.

---

## Decision 5 — CLI Before APIs

CLI providers already exist and solve real developer workflows today.

API support can be added later through the Adapter framework.

---

# 23.28 Architectural Observation

Version 0.1 deliberately leaves many ambitious ideas unimplemented.

This is a feature, not a limitation.

The architecture is designed so that:

* semantic search can be added without changing Memory,
* APIs can be added without changing Workflows,
* MCP can be added without changing Context Builder,
* cloud synchronization can be added without changing Storage.

This separation allows Context OS to evolve incrementally while preserving a stable foundation.

---

# 23.29 Chapter Summary

This chapter defines the precise scope of Context OS Version 0.1.

Rather than attempting to become a complete autonomous engineering platform, Version 0.1 focuses on validating a single architectural hypothesis:

> **Project intelligence should outlive conversations and remain independent of AI providers.**

By delivering durable workflows, provider-agnostic context reconstruction, checkpoint recovery, and human-readable project memory, Version 0.1 establishes the foundation upon which every future capability described in this document can be built.

The next chapter presents the **Roadmap**, outlining how Context OS evolves from Version 0.1 into a mature engineering platform through carefully staged releases while preserving backward compatibility and architectural consistency.
