package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestDiscover_FindsRootDirectly(t *testing.T) {
	dir := t.TempDir()
	if err := project.CreateLayout(dir); err != nil {
		t.Fatal(err)
	}

	got, err := project.Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != dir {
		t.Errorf("Discover() = %q, want %q", got, dir)
	}
}

func TestDiscover_FindsRootFromSubdirectory(t *testing.T) {
	root := t.TempDir()
	if err := project.CreateLayout(root); err != nil {
		t.Fatal(err)
	}

	// Create a nested subdirectory and start discovery from there.
	sub := filepath.Join(root, "internal", "pkg", "service")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := project.Discover(sub)
	if err != nil {
		t.Fatalf("Discover from subdir: %v", err)
	}
	if got != root {
		t.Errorf("Discover() = %q, want %q", got, root)
	}
}

func TestDiscover_NotFound(t *testing.T) {
	// Use a temp dir with no .context/ at any level.
	dir := t.TempDir()

	_, err := project.Discover(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeNotFound {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeNotFound)
	}
}

func TestDiscover_StopsAtProjectBoundary(t *testing.T) {
	// Outer project
	outer := t.TempDir()
	if err := project.CreateLayout(outer); err != nil {
		t.Fatal(err)
	}

	// Inner project nested inside outer
	inner := filepath.Join(outer, "nested-project")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := project.CreateLayout(inner); err != nil {
		t.Fatal(err)
	}

	// Discover from inside inner — should return inner, not outer.
	sub := filepath.Join(inner, "src")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := project.Discover(sub)
	if err != nil {
		t.Fatal(err)
	}
	if got != inner {
		t.Errorf("Discover() = %q, want inner project %q", got, inner)
	}
}
