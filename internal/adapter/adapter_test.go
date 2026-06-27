package adapter_test

import (
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
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}
	if adapter.IsInjected(dir, p) {
		t.Error("expected IsInjected=false for missing file")
	}
}

func TestIsInjected_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# existing content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if adapter.IsInjected(dir, p) {
		t.Error("expected IsInjected=false for file without marker")
	}
}

func TestInject_AppendProvider(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}

	// Inject into non-existent file (should create it).
	if err := adapter.Inject(dir, p); err != nil {
		t.Fatalf("Inject: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<!-- context-os -->") {
		t.Error("injected file missing marker")
	}
	if !strings.Contains(string(data), "context workflow start") {
		t.Error("injected file missing usage text")
	}
}

func TestInject_Idempotent(t *testing.T) {
	dir := t.TempDir()
	p := adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"}

	if err := adapter.Inject(dir, p); err != nil {
		t.Fatal(err)
	}
	firstContent, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))

	// Second inject must not change the file.
	if err := adapter.Inject(dir, p); err != nil {
		t.Fatal(err)
	}
	secondContent, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))

	if string(firstContent) != string(secondContent) {
		t.Error("Inject changed file on second call (not idempotent)")
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

	path := filepath.Join(dir, ".cursor", "rules", "context-os.mdc")
	data, err := os.ReadFile(path)
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

	_, err := os.Stat(filepath.Join(dir, ".github"))
	if err != nil {
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
