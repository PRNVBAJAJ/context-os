package application_test

import (
	"context"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/application"
)

func TestGetProjectStatus_Success(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	// Initialize a project first.
	created, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "status-test",
		RootPath: dir,
		Language: "go",
	})
	if err != nil {
		t.Fatalf("InitializeProject: %v", err)
	}

	status, err := application.GetProjectStatus(ctx, dir)
	if err != nil {
		t.Fatalf("GetProjectStatus: %v", err)
	}

	if status.Project == nil {
		t.Fatal("status.Project is nil")
	}
	if status.Project.ID != created.ID {
		t.Errorf("ID = %q, want %q", status.Project.ID, created.ID)
	}
	if status.Project.Name != "status-test" {
		t.Errorf("Name = %q, want %q", status.Project.Name, "status-test")
	}
	if status.Project.Language != "go" {
		t.Errorf("Language = %q, want %q", status.Project.Language, "go")
	}
}

func TestGetProjectStatus_NotInitialized(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	_, err := application.GetProjectStatus(ctx, dir)
	if err == nil {
		t.Fatal("expected error for uninitialized project, got nil")
	}
}
