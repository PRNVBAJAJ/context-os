package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

func TestStartWorkflow_Success(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "wf-test",
		RootPath: dir,
	}); err != nil {
		t.Fatalf("InitializeProject: %v", err)
	}

	w, err := application.StartWorkflow(ctx, application.StartWorkflowOptions{
		RootPath:    dir,
		Name:        "implement auth",
		Description: "Add JWT-based auth",
	})
	if err != nil {
		t.Fatalf("StartWorkflow: %v", err)
	}
	if w.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
	if w.Name != "implement auth" {
		t.Errorf("Name = %q, want %q", w.Name, "implement auth")
	}
	if w.Status != workflow.StatusRunning {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusRunning)
	}
	if w.StartedAt == nil {
		t.Error("StartedAt should be set")
	}
}

func TestStartWorkflow_EmitsEvent(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "event-wf-test",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := application.StartWorkflow(ctx, application.StartWorkflowOptions{
		RootPath: dir,
		Name:     "implement auth",
	}); err != nil {
		t.Fatal(err)
	}

	// RunDoctor counts all events: project.initialized + workflow.started = 2.
	result, err := application.RunDoctor(ctx, dir)
	if err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}
	if result.EventCount != 2 {
		t.Errorf("EventCount = %d, want 2 (project.initialized + workflow.started)", result.EventCount)
	}
}

func TestListWorkflows_ReturnsAllWorkflows(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "list-wf-test",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	names := []string{"first", "second", "third"}
	for _, name := range names {
		if _, err := application.StartWorkflow(ctx, application.StartWorkflowOptions{
			RootPath: dir,
			Name:     name,
		}); err != nil {
			t.Fatalf("StartWorkflow(%q): %v", name, err)
		}
	}

	workflows, err := application.ListWorkflows(ctx, application.ListWorkflowsOptions{RootPath: dir})
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(workflows) != 3 {
		t.Fatalf("ListWorkflows returned %d, want 3", len(workflows))
	}
}

func initProjectWithWorkflow(t *testing.T) (dir string, wfID string) {
	t.Helper()
	dir = t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "trans-test", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}
	w, err := application.StartWorkflow(ctx, application.StartWorkflowOptions{
		RootPath: dir, Name: "implement auth",
	})
	if err != nil {
		t.Fatal(err)
	}
	return dir, w.ID.String()[:8]
}

func TestCompleteWorkflow_Success(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	w, err := application.CompleteWorkflow(ctx, application.CompleteWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	})
	if err != nil {
		t.Fatalf("CompleteWorkflow: %v", err)
	}
	if w.Status != workflow.StatusCompleted {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusCompleted)
	}
	if w.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestFailWorkflow_Success(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	w, err := application.FailWorkflow(ctx, application.FailWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	})
	if err != nil {
		t.Fatalf("FailWorkflow: %v", err)
	}
	if w.Status != workflow.StatusFailed {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusFailed)
	}
}

func TestPauseAndResumeWorkflow(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	paused, err := application.PauseWorkflow(ctx, application.PauseWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	})
	if err != nil {
		t.Fatalf("PauseWorkflow: %v", err)
	}
	if paused.Status != workflow.StatusPaused {
		t.Errorf("after Pause, Status = %q, want %q", paused.Status, workflow.StatusPaused)
	}

	resumed, err := application.ResumeWorkflow(ctx, application.ResumeWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	})
	if err != nil {
		t.Fatalf("ResumeWorkflow: %v", err)
	}
	if resumed.Status != workflow.StatusRunning {
		t.Errorf("after Resume, Status = %q, want %q", resumed.Status, workflow.StatusRunning)
	}
}

func TestDeleteWorkflow_CompletedAllowed(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	if _, err := application.CompleteWorkflow(ctx, application.CompleteWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	}); err != nil {
		t.Fatal(err)
	}

	if err := application.DeleteWorkflow(ctx, application.DeleteWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	}); err != nil {
		t.Fatalf("DeleteWorkflow: %v", err)
	}

	workflows, _ := application.ListWorkflows(ctx, application.ListWorkflowsOptions{RootPath: dir})
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows after delete, got %d", len(workflows))
	}
}

func TestDeleteWorkflow_FailedAllowed(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	if _, err := application.FailWorkflow(ctx, application.FailWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	}); err != nil {
		t.Fatal(err)
	}

	if err := application.DeleteWorkflow(ctx, application.DeleteWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	}); err != nil {
		t.Fatalf("DeleteWorkflow on failed: %v", err)
	}
}

func TestDeleteWorkflow_RunningRejected(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	err := application.DeleteWorkflow(ctx, application.DeleteWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	})
	if err == nil {
		t.Fatal("expected error deleting running workflow")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) || domainErr.Code != shared.CodeInvalidInput {
		t.Errorf("want CodeInvalidInput, got %v", err)
	}
}

func TestDeleteWorkflow_PausedRejected(t *testing.T) {
	dir, prefix := initProjectWithWorkflow(t)
	ctx := context.Background()

	if _, err := application.PauseWorkflow(ctx, application.PauseWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	}); err != nil {
		t.Fatal(err)
	}

	err := application.DeleteWorkflow(ctx, application.DeleteWorkflowOptions{
		RootPath: dir, IDPrefix: prefix,
	})
	if err == nil {
		t.Fatal("expected error deleting paused workflow")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) || domainErr.Code != shared.CodeInvalidInput {
		t.Errorf("want CodeInvalidInput, got %v", err)
	}
}

func TestTransition_InvalidPrefix(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name: "no-wf-project", RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	_, err := application.CompleteWorkflow(ctx, application.CompleteWorkflowOptions{
		RootPath: dir, IDPrefix: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent prefix, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeNotFound {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeNotFound)
	}
}

func TestListWorkflows_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "empty-wf",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	workflows, err := application.ListWorkflows(ctx, application.ListWorkflowsOptions{RootPath: dir})
	if err != nil {
		t.Fatalf("ListWorkflows: %v", err)
	}
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows, got %d", len(workflows))
	}
}
