package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := project.CreateLayout(dir); err != nil {
		t.Fatal(err)
	}

	original := &project.Project{
		ID:             shared.NewID(),
		Name:           "round-trip",
		RootPath:       dir,
		Language:       "go",
		RuntimeVersion: shared.Version,
		SchemaVersion:  0,
		CreatedAt:      time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC),
	}

	if err := project.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := project.Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.ID != original.ID {
		t.Errorf("ID = %q, want %q", got.ID, original.ID)
	}
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
	if got.Language != original.Language {
		t.Errorf("Language = %q, want %q", got.Language, original.Language)
	}
	if got.RootPath != dir {
		t.Errorf("RootPath = %q, want %q", got.RootPath, dir)
	}
	if got.RuntimeVersion != original.RuntimeVersion {
		t.Errorf("RuntimeVersion = %q, want %q", got.RuntimeVersion, original.RuntimeVersion)
	}
	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, original.CreatedAt)
	}
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := project.Load(dir)
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

func TestSave_RootPathNotInFile(t *testing.T) {
	dir := t.TempDir()
	if err := project.CreateLayout(dir); err != nil {
		t.Fatal(err)
	}

	p, err := project.New("portability-test", dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := project.Save(p); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(filepath.Join(project.Dir(dir), "project.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// root_path must not be written to disk — it is derived at runtime from location.
	if strings.Contains(string(raw), "root_path") {
		t.Errorf("project.yaml should not contain root_path field:\n%s", raw)
	}
}
