package checkpoint_test

import (
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/checkpoint"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestNew_WithWorkflow(t *testing.T) {
	wfID := shared.NewID()
	cp := checkpoint.New(wfID, "before refactor")

	if cp.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
	if cp.WorkflowID != wfID {
		t.Errorf("WorkflowID = %q, want %q", cp.WorkflowID, wfID)
	}
	if cp.Note != "before refactor" {
		t.Errorf("Note = %q, want %q", cp.Note, "before refactor")
	}
	if cp.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestNew_ProjectLevel(t *testing.T) {
	cp := checkpoint.New(shared.EmptyID, "")
	if !cp.WorkflowID.IsEmpty() {
		t.Errorf("WorkflowID should be empty for project-level checkpoint, got %q", cp.WorkflowID)
	}
}

func TestNew_EmptyNoteIsValid(t *testing.T) {
	wfID := shared.NewID()
	cp := checkpoint.New(wfID, "")
	if cp.ID.IsEmpty() {
		t.Error("ID should not be empty even with empty note")
	}
}
