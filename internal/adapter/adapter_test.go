package adapter_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/adapter"
)

func TestKnownProviders(t *testing.T) {
	providers := adapter.KnownProviders()
	if len(providers) == 0 {
		t.Fatal("KnownProviders returned empty slice")
	}
	names := make(map[string]bool)
	for _, p := range providers {
		if p.Name == "" || p.Binary == "" || p.ConfigPath == "" {
			t.Errorf("provider %+v has empty required field", p)
		}
		names[p.Name] = true
	}
	for _, want := range []string{"claude", "cursor", "copilot", "gemini", "opencode", "codex", "aider"} {
		if !names[want] {
			t.Errorf("missing provider %q", want)
		}
	}
}

func TestIsInjected_NotPresent(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "gemini", Binary: "gemini", ConfigPath: "GEMINI.md"}
	if adapter.IsInjected(dir, p) {
		t.Error("expected IsInjected=false for missing file")
	}
}

func TestIsInjected_FileWithoutMarker(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "gemini", Binary: "gemini", ConfigPath: "GEMINI.md"}
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("# existing content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if adapter.IsInjected(dir, p) {
		t.Error("expected IsInjected=false for file without marker")
	}
}

// Claude provider: IsInjected requires BOTH CLAUDE.md marker AND .claude/settings.json hook.

func TestIsInjected_Claude_RequiresBothMarkerAndHook(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}

	// Only CLAUDE.md marker — hook missing → not injected.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("<!-- context-os -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if adapter.IsInjected(dir, p) {
		t.Error("expected IsInjected=false when hook is missing")
	}

	// Now inject fully (writes both CLAUDE.md body + hook).
	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject: %v", err)
	}
	if !adapter.IsInjected(dir, p) {
		t.Error("expected IsInjected=true after full inject")
	}
}

func TestInject_Claude_WritesHookFile(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf(".claude/settings.json not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "context status") {
		t.Error("hook command missing from settings.json")
	}
	if !strings.Contains(content, "UserPromptSubmit") {
		t.Error("UserPromptSubmit key missing from settings.json")
	}

	// Must be valid JSON.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Errorf("settings.json is not valid JSON: %v", err)
	}
}

func TestInject_Claude_HookMergesExistingSettings(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Existing settings with an unrelated key.
	existing := `{"permissions": {"allow": ["Bash(git:*)"]}}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, ".claude", "settings.json"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}
	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	content := string(data)

	// Existing key must be preserved.
	if !strings.Contains(content, "permissions") {
		t.Error("existing 'permissions' key was lost during merge")
	}
	// Hook must be added.
	if !strings.Contains(content, "context status") {
		t.Error("hook command missing after merge")
	}
}

func TestInject_Claude_Idempotent(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatal(err)
	}
	claudeMdFirst, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	settingsFirst, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatal(err)
	}
	claudeMdSecond, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	settingsSecond, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))

	if string(claudeMdFirst) != string(claudeMdSecond) {
		t.Error("CLAUDE.md changed on second inject")
	}
	if string(settingsFirst) != string(settingsSecond) {
		t.Error("settings.json changed on second inject")
	}
}

func TestInject_AppendToExisting(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "gemini", Binary: "gemini", ConfigPath: "GEMINI.md"}
	existing := "# My Gemini Config\n\nSome existing instructions.\n"
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "GEMINI.md"))
	content := string(data)
	if !strings.HasPrefix(content, existing) {
		t.Error("existing content was modified or overwritten")
	}
	if !strings.Contains(content, "<!-- context-os -->") {
		t.Error("marker not found after append")
	}
}

func TestInject_CursorCreatesDedicatedFile(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{
		Name:       "cursor",
		Binary:     "cursor",
		ConfigPath: filepath.Join(".cursor", "rules", "context-os.mdc"),
		NeedsDir:   true,
	}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject cursor: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".cursor", "rules", "context-os.mdc"))
	if err != nil {
		t.Fatalf("cursor file not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "alwaysApply: true") {
		t.Error("cursor file missing frontmatter")
	}
	if !strings.Contains(content, "<!-- context-os -->") {
		t.Error("cursor file missing marker")
	}
}

func TestInject_NeedsDir(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{
		Name:       "copilot",
		Binary:     "gh",
		ConfigPath: filepath.Join(".github", "copilot-instructions.md"),
		NeedsDir:   true,
	}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject copilot: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".github")); err != nil {
		t.Error(".github directory was not created")
	}
}

func TestDetect_ReturnsAllProviders(t *testing.T) {
	dir := t.TempDir()
	results := adapter.Detect(dir)
	if len(results) != len(adapter.KnownProviders()) {
		t.Errorf("Detect returned %d results, want %d", len(results), len(adapter.KnownProviders()))
	}
	for _, r := range results {
		if r.Provider.Name == "" {
			t.Error("result has empty provider name")
		}
	}
}

func TestInject_Claude_AddsClaudeToGitignore(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	if !strings.Contains(string(data), ".claude/") {
		t.Error(".gitignore does not contain .claude/ entry")
	}
}

func TestInject_Claude_GitignoreIdempotent(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing .gitignore that already excludes .claude/.
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules/\n.claude/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}
	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	count := strings.Count(string(data), ".claude/")
	if count != 1 {
		t.Errorf(".claude/ appears %d times in .gitignore, want 1", count)
	}
}

func TestDetect_Claude_NotInjectedUntilHookPresent(t *testing.T) {
	dir := t.TempDir()
	// Write CLAUDE.md with marker but no hook.
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("<!-- context-os -->\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	results := adapter.Detect(dir)
	for _, r := range results {
		if r.Provider.Name == "claude" && r.Injected {
			t.Error("claude should not be Injected=true when hook is missing")
		}
	}
}
