package application_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestInitializeProject_Success(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	p, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "test-project",
		RootPath: dir,
		Language: "go",
	})

	if err != nil {
		t.Fatalf("InitializeProject: %v", err)
	}
	if p == nil {
		t.Fatal("returned project is nil")
	}
	if p.ID.IsEmpty() {
		t.Error("project ID must not be empty")
	}
	if p.Name != "test-project" {
		t.Errorf("Name = %q, want %q", p.Name, "test-project")
	}
	if p.RootPath != dir {
		t.Errorf("RootPath = %q, want %q", p.RootPath, dir)
	}
}

func TestInitializeProject_CreatesLayout(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	expectedPaths := []string{
		".context",
		".context/project.yaml",
		".context/runtime.db",
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

	for _, rel := range expectedPaths {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("expected %q to exist: %v", rel, err)
		}
	}
}

func TestInitializeProject_NameDerivedFromPath(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "my-app")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	p, err := application.InitializeProject(context.Background(), application.InitOptions{
		RootPath: dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "my-app" {
		t.Errorf("Name = %q, want %q", p.Name, "my-app")
	}
}

func TestInitializeProject_ConflictOnDoubleInit(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	opts := application.InitOptions{RootPath: dir}

	if _, err := application.InitializeProject(ctx, opts); err != nil {
		t.Fatalf("first init: %v", err)
	}

	_, err := application.InitializeProject(ctx, opts)
	if err == nil {
		t.Fatal("expected error on second init, got nil")
	}

	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeConflict {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeConflict)
	}
}

func TestInitializeProject_ProjectYamlIsLoadable(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	created, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "yaml-loadable",
		RootPath: dir,
		Language: "python",
	})
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ID != created.ID {
		t.Errorf("loaded ID = %q, want %q", loaded.ID, created.ID)
	}
	if loaded.Language != "python" {
		t.Errorf("loaded Language = %q, want %q", loaded.Language, "python")
	}
}
