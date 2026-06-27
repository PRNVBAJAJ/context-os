package adapter

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	injectionMarker    = "<!-- context-os -->"
	injectionEndMarker = "<!-- end context-os -->"
)

// injectionBody is the Context OS usage block appended to provider config files.
const injectionBody = `
<!-- context-os -->
## Context OS

This project uses Context OS for durable workflow state across AI sessions.

Before starting any multi-step task:
  context workflow start --name "<task name>"

At meaningful milestones:
  context checkpoint create --note "<what's done and what's next>"

For non-obvious decisions or constraints:
  context memory add <slug> "<the insight>"

At the start of a new session, orient first:
  context status && context workflow list && context memory list
<!-- end context-os -->
`

// cursorFileContent is the full content of .cursor/rules/context-os.mdc.
const cursorFileContent = `---
description: Context OS workflow integration
alwaysApply: true
---

<!-- context-os -->
## Context OS

This project uses Context OS for durable workflow state across AI sessions.

Before starting any multi-step task:
  context workflow start --name "<task name>"

At meaningful milestones:
  context checkpoint create --note "<what's done and what's next>"

For non-obvious decisions or constraints:
  context memory add <slug> "<the insight>"

At the start of a new session, orient first:
  context status && context workflow list && context memory list
<!-- end context-os -->
`

// Provider describes a known AI CLI tool and where it reads project instructions.
type Provider struct {
	Name       string // e.g. "claude", "cursor"
	Binary     string // executable name searched via exec.LookPath
	ConfigPath string // path relative to project root
	NeedsDir   bool   // true if the parent directory must be created
}

// DetectionResult is the runtime state of a provider for a specific project root.
type DetectionResult struct {
	Provider Provider
	Detected bool // binary found in PATH
	Injected bool // Context OS block already present in config file
}

// KnownProviders returns the full registry of supported AI CLI tools.
// AGENTS.md is listed once for opencode; codex reuses the same path.
func KnownProviders() []Provider {
	return []Provider{
		{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md", NeedsDir: false},
		{Name: "cursor", Binary: "cursor", ConfigPath: filepath.Join(".cursor", "rules", "context-os.mdc"), NeedsDir: true},
		{Name: "copilot", Binary: "gh", ConfigPath: filepath.Join(".github", "copilot-instructions.md"), NeedsDir: true},
		{Name: "gemini", Binary: "gemini", ConfigPath: "GEMINI.md", NeedsDir: false},
		{Name: "opencode", Binary: "opencode", ConfigPath: "AGENTS.md", NeedsDir: false},
		{Name: "codex", Binary: "codex", ConfigPath: "AGENTS.md", NeedsDir: false},
		{Name: "aider", Binary: "aider", ConfigPath: "CONVENTIONS.md", NeedsDir: false},
	}
}

// Detect checks which providers are installed and whether their config files
// already contain the Context OS block.
func Detect(rootPath string) []DetectionResult {
	providers := KnownProviders()
	results := make([]DetectionResult, 0, len(providers))
	for _, p := range providers {
		_, err := exec.LookPath(p.Binary)
		results = append(results, DetectionResult{
			Provider: p,
			Detected: err == nil,
			Injected: IsInjected(rootPath, p),
		})
	}
	return results
}

// IsInjected reports whether the provider's config file in rootPath already
// contains the Context OS injection marker.
func IsInjected(rootPath string, p Provider) bool {
	path := filepath.Join(rootPath, p.ConfigPath)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), injectionMarker)
}

// Inject writes the Context OS usage block to the provider's config file.
// It is idempotent: if the marker is already present the file is unchanged.
// The parent directory is created when p.NeedsDir is true.
func Inject(rootPath string, p Provider) error {
	if IsInjected(rootPath, p) {
		return nil
	}

	path := filepath.Join(rootPath, p.ConfigPath)

	if p.NeedsDir {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}

	// Cursor gets its own standalone file; all other providers append.
	if p.Name == "cursor" {
		return os.WriteFile(path, []byte(cursorFileContent), 0o644)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = f.WriteString(injectionBody)
	return err
}
