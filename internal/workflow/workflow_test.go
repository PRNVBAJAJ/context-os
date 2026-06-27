package workflow_test

import (
	"errors"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

func TestNew_ValidInputs(t *testing.T) {
	w, err := workflow.New("implement auth", "Add JWT-based authentication")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if w.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
	if w.Name != "implement auth" {
		t.Errorf("Name = %q, want %q", w.Name, "implement auth")
	}
	if w.Description != "Add JWT-based authentication" {
		t.Errorf("Description = %q, want %q", w.Description, "Add JWT-based authentication")
	}
	if w.Status != workflow.StatusPending {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusPending)
	}
	if w.StartedAt != nil {
		t.Error("StartedAt should be nil for a new workflow")
	}
	if w.CompletedAt != nil {
		t.Error("CompletedAt should be nil for a new workflow")
	}
}

func TestNew_InvalidName(t *testing.T) {
	cases := []struct{ name, desc string }{
		{"", "empty name"},
		{"   ", "whitespace-only name"},
		{"impl/auth", "name with forward slash"},
		{"impl\\auth", "name with backslash"},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := workflow.New(tc.name, "")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var domainErr *shared.Error
			if !errors.As(err, &domainErr) {
				t.Fatalf("expected *shared.Error, got %T", err)
			}
			if domainErr.Code != shared.CodeInvalidInput {
				t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeInvalidInput)
			}
		})
	}
}

func TestWorkflow_Start(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	if err := w.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if w.Status != workflow.StatusRunning {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusRunning)
	}
	if w.StartedAt == nil {
		t.Error("StartedAt should be set after Start()")
	}
}

func TestWorkflow_Start_OnlyFromPending(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	w.Start() //nolint:errcheck

	err := w.Start() // already running
	if err == nil {
		t.Fatal("expected error starting a running workflow, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) || domainErr.Code != shared.CodeInvalidInput {
		t.Errorf("expected CodeInvalidInput, got %v", err)
	}
}

func TestWorkflow_Complete(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	w.Start() //nolint:errcheck

	if err := w.Complete(); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if w.Status != workflow.StatusCompleted {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusCompleted)
	}
	if w.CompletedAt == nil {
		t.Error("CompletedAt should be set after Complete()")
	}
}

func TestWorkflow_Complete_OnlyFromRunning(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	err := w.Complete() // still pending
	if err == nil {
		t.Fatal("expected error completing a pending workflow, got nil")
	}
}

func TestWorkflow_Fail(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	w.Start() //nolint:errcheck

	if err := w.Fail(); err != nil {
		t.Fatalf("Fail: %v", err)
	}
	if w.Status != workflow.StatusFailed {
		t.Errorf("Status = %q, want %q", w.Status, workflow.StatusFailed)
	}
	if w.CompletedAt == nil {
		t.Error("CompletedAt should be set after Fail()")
	}
}

func TestWorkflow_PauseAndResume(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	w.Start() //nolint:errcheck

	if err := w.Pause(); err != nil {
		t.Fatalf("Pause: %v", err)
	}
	if w.Status != workflow.StatusPaused {
		t.Errorf("after Pause, Status = %q, want %q", w.Status, workflow.StatusPaused)
	}

	if err := w.Resume(); err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if w.Status != workflow.StatusRunning {
		t.Errorf("after Resume, Status = %q, want %q", w.Status, workflow.StatusRunning)
	}
}

func TestWorkflow_Pause_OnlyFromRunning(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	if err := w.Pause(); err == nil {
		t.Error("expected error pausing a pending workflow, got nil")
	}
}

func TestWorkflow_Resume_OnlyFromPaused(t *testing.T) {
	w, _ := workflow.New("implement auth", "")
	w.Start() //nolint:errcheck
	if err := w.Resume(); err == nil {
		t.Error("expected error resuming a running workflow, got nil")
	}
}
