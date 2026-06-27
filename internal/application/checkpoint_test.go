package application_test

import (
	"context"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestCreateCheckpoint_ProjectLevel(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "cp-test", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	cp, err := application.CreateCheckpoint(ctx, application.CreateCheckpointOptions{
		RootPath: dir,
		Note:     "before refactor",
	})
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}
	if cp.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
	if !cp.WorkflowID.IsEmpty() {
		t.Errorf("WorkflowID should be empty for project-level checkpoint, got %q", cp.WorkflowID)
	}
	if cp.Note != "before refactor" {
		t.Errorf("Note = %q, want %q", cp.Note, "before refactor")
	}
}

func TestCreateCheckpoint_WithWorkflow(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "cp-wf-test", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	w, err := application.StartWorkflow(ctx, application.StartWorkflowOptions{
		RootPath: dir, Name: "implement auth",
	})
	if err != nil {
		t.Fatal(err)
	}

	cp, err := application.CreateCheckpoint(ctx, application.CreateCheckpointOptions{
		RootPath:         dir,
		WorkflowIDPrefix: w.ID.String()[:8],
		Note:             "mid-implementation",
	})
	if err != nil {
		t.Fatalf("CreateCheckpoint with workflow: %v", err)
	}
	if cp.WorkflowID != w.ID {
		t.Errorf("WorkflowID = %q, want %q", cp.WorkflowID, w.ID)
	}
}

func TestCreateCheckpoint_InvalidWorkflowPrefix(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "cp-bad-prefix", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := application.CreateCheckpoint(ctx, application.CreateCheckpointOptions{
		RootPath:         dir,
		WorkflowIDPrefix: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for invalid prefix, got nil")
	}
}

func TestListCheckpoints(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "list-cp-test", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	for _, note := range []string{"first", "second", "third"} {
		if _, err := application.CreateCheckpoint(ctx, application.CreateCheckpointOptions{
			RootPath: dir, Note: note,
		}); err != nil {
			t.Fatalf("CreateCheckpoint(%q): %v", note, err)
		}
	}

	checkpoints, err := application.ListCheckpoints(ctx, application.ListCheckpointsOptions{RootPath: dir})
	if err != nil {
		t.Fatalf("ListCheckpoints: %v", err)
	}
	if len(checkpoints) != 3 {
		t.Errorf("ListCheckpoints returned %d, want 3", len(checkpoints))
	}
}

func TestListCheckpoints_Empty(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "empty-cp", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	checkpoints, err := application.ListCheckpoints(ctx, application.ListCheckpointsOptions{RootPath: dir})
	if err != nil {
		t.Fatalf("ListCheckpoints: %v", err)
	}
	if len(checkpoints) != 0 {
		t.Errorf("expected 0 checkpoints, got %d", len(checkpoints))
	}
}

func TestCreateCheckpoint_EmptyNoteIsValid(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "no-note", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	cp, err := application.CreateCheckpoint(ctx, application.CreateCheckpointOptions{
		RootPath: dir,
	})
	if err != nil {
		t.Fatalf("CreateCheckpoint with empty note: %v", err)
	}
	if cp.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
	_ = shared.EmptyID // imported for completeness
}
