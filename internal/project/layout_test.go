package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestDir(t *testing.T) {
	root := filepath.Join("/home", "user", "my-repo")
	got := project.Dir(root)
	want := filepath.Join(root, project.ContextDir)
	if got != want {
		t.Errorf("Dir() = %q, want %q", got, want)
	}
}

func TestIsInitialized(t *testing.T) {
	dir := t.TempDir()

	if project.IsInitialized(dir) {
		t.Error("should not be initialized before CreateLayout")
	}

	if err := os.Mkdir(filepath.Join(dir, project.ContextDir), 0o755); err != nil {
		t.Fatal(err)
	}

	if !project.IsInitialized(dir) {
		t.Error("should be initialized after creating .context/")
	}
}

func TestCreateLayout(t *testing.T) {
	dir := t.TempDir()

	if err := project.CreateLayout(dir); err != nil {
		t.Fatalf("CreateLayout returned error: %v", err)
	}

	expectedDirs := []string{
		".context",
		".context/workflows",
		".context/sessions",
		".context/memory",
		".context/artifacts",
		".context/checkpoints",
		".context/providers",
		".context/events",
		".context/logs",
		".context/cache",
		".context/temp",
		".context/plugins",
	}

	for _, rel := range expectedDirs {
		path := filepath.Join(dir, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected directory %q to exist: %v", rel, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q should be a directory", rel)
		}
	}
}

func TestCreateLayout_ConflictOnSecondCall(t *testing.T) {
	dir := t.TempDir()

	if err := project.CreateLayout(dir); err != nil {
		t.Fatalf("first CreateLayout: %v", err)
	}

	err := project.CreateLayout(dir)
	if err == nil {
		t.Fatal("expected error on second CreateLayout, got nil")
	}

	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeConflict {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeConflict)
	}
}
