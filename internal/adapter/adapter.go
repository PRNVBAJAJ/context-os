package adapter

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	injectionMarker = "<!-- context-os -->"

	// claudeSettingsPath is the project-level Claude Code settings file.
	// Committing this file gives every collaborator the same hooks automatically.
	claudeSettingsPath = ".claude/settings.json"

	// claudeHookScript is injected as a UserPromptSubmit hook in .claude/settings.json.
	// When a workflow is running it surfaces project state; when none is active it
	// emits a hard requirement so the AI cannot proceed without starting one.
	claudeHookScript = `if context workflow list 2>/dev/null | grep -q running; then context status 2>/dev/null && echo '---' && context workflow list 2>/dev/null && echo '---' && context memory list 2>/dev/null; else echo '⚠ NO ACTIVE WORKFLOW. You MUST run: context workflow start "<task name>" before doing anything.'; fi || true`

	// claudeStopScript is injected as a Stop hook in .claude/settings.json.
	// It auto-checkpoints after every Claude turn when a workflow is running,
	// removing the AI's discretion over whether to checkpoint.
	claudeStopScript = `context workflow list 2>/dev/null | grep -q running && context checkpoint create --note 'auto' 2>/dev/null || true`

	// claudePostToolUseScript is injected as a PostToolUse hook in .claude/settings.json.
	// It silently records every file touched by Read/Edit/Write tools so that
	// `context status` can surface hot files without manual tracking.
	claudePostToolUseScript = `cat | context track 2>/dev/null || true`
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

type claudeHookEntry struct {
	Matcher string       `json:"matcher"`
	Hooks   []claudeHook `json:"hooks"`
}

type claudeHook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

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
	Injected bool // Context OS fully configured for this provider
}

// KnownProviders returns the full registry of supported AI CLI tools.
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

// Detect checks which providers are installed and whether Context OS is fully
// configured for each one.
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

// IsInjected reports whether Context OS is fully configured for the provider.
// For "claude" this means both CLAUDE.md has the marker AND .claude/settings.json
// contains the UserPromptSubmit hook.
func IsInjected(rootPath string, p Provider) bool {
	data, err := os.ReadFile(filepath.Join(rootPath, p.ConfigPath))
	if err != nil || !strings.Contains(string(data), injectionMarker) {
		return false
	}
	if p.Name == "claude" {
		return isClaudeHookInjected(rootPath)
	}
	return true
}

// Inject configures Context OS for the given provider. It is fully idempotent:
// each step only runs if that step is not already complete.
//
// For "claude" this means:
//  1. Append the usage block to CLAUDE.md (if not present)
//  2. Merge the UserPromptSubmit hook into .claude/settings.json (if not present)
//
// For "cursor" a standalone .cursor/rules/context-os.mdc is written.
// All other providers append the usage block to their config file.
func Inject(rootPath string, p Provider) error {
	configPath := filepath.Join(rootPath, p.ConfigPath)

	if !isConfigMarkerPresent(configPath) {
		if err := injectConfigFile(configPath, p); err != nil {
			return err
		}
	}

	if p.Name == "claude" {
		_ = injectClaudeHook(rootPath)
	}

	return nil
}

// isConfigMarkerPresent reports whether the config file already contains the marker.
func isConfigMarkerPresent(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && strings.Contains(string(data), injectionMarker)
}

// injectConfigFile writes the injection body to the provider's config file.
func injectConfigFile(path string, p Provider) error {
	if p.NeedsDir {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
	}
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

// isClaudeHookInjected reports whether .claude/settings.json contains all three
// Context OS hooks: UserPromptSubmit, Stop, and PostToolUse.
func isClaudeHookInjected(rootPath string) bool {
	data, err := os.ReadFile(filepath.Join(rootPath, claudeSettingsPath))
	if err != nil {
		return false
	}
	s := string(data)
	return strings.Contains(s, "NO ACTIVE WORKFLOW") &&
		strings.Contains(s, "checkpoint create") &&
		strings.Contains(s, "context track")
}

// injectClaudeHook writes or merges the UserPromptSubmit and Stop hooks into
// .claude/settings.json, preserving any existing settings.
// It also ensures .claude/ is listed in the project's .gitignore so that
// the file is never accidentally committed.
func injectClaudeHook(rootPath string) error {
	if isClaudeHookInjected(rootPath) {
		return nil
	}

	settingsPath := filepath.Join(rootPath, claudeSettingsPath)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	// Parse existing settings into a raw map so unknown fields are preserved.
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &raw)
	}

	// Parse existing hooks map.
	hooksMap := make(map[string][]claudeHookEntry)
	if raw["hooks"] != nil {
		_ = json.Unmarshal(raw["hooks"], &hooksMap)
	}

	// Ensure all three Context OS hooks are present, replacing older entries.
	hooksMap["UserPromptSubmit"] = upsertContextHook(hooksMap["UserPromptSubmit"], claudeHookScript)
	hooksMap["Stop"] = upsertContextHook(hooksMap["Stop"], claudeStopScript)
	hooksMap["PostToolUse"] = upsertContextHook(hooksMap["PostToolUse"], claudePostToolUseScript)

	hooksBytes, err := json.Marshal(hooksMap)
	if err != nil {
		return err
	}
	raw["hooks"] = json.RawMessage(hooksBytes)

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0o644); err != nil {
		return err
	}

	// Keep .claude/ out of version control. It contains per-developer settings
	// that should not be shared — each developer gets their own hook on init.
	_ = ensureGitignored(rootPath, ".claude/")
	return nil
}

// upsertContextHook ensures exactly one Context OS entry exists for command.
// If command is already present, it's a no-op. If an older Context OS entry
// (identified by a "context " prefix) is found, it is replaced in-place.
// Otherwise the entry is appended.
func upsertContextHook(entries []claudeHookEntry, command string) []claudeHookEntry {
	for i, e := range entries {
		for j, h := range e.Hooks {
			if h.Command == command {
				return entries // already up to date
			}
			if strings.HasPrefix(h.Command, "context ") || strings.HasPrefix(h.Command, "if context ") {
				// Old Context OS entry — upgrade in place.
				entries[i].Hooks[j].Command = command
				return entries
			}
		}
	}
	return append(entries, claudeHookEntry{
		Matcher: "",
		Hooks:   []claudeHook{{Type: "command", Command: command}},
	})
}

// ensureGitignored appends pattern to <rootPath>/.gitignore if not already present.
func ensureGitignored(rootPath, pattern string) error {
	gitignorePath := filepath.Join(rootPath, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == pattern {
				return nil // already present
			}
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// Ensure the entry starts on its own line.
	entry := pattern + "\n"
	if len(data) > 0 && data[len(data)-1] != '\n' {
		entry = "\n" + entry
	}
	_, err = f.WriteString(entry)
	return err
}
