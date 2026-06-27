package workflow

import (
	"strings"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// Status represents the lifecycle stage of a workflow.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Workflow is the core unit of durable work in Context OS.
// It tracks a named engineering task from start to completion across
// provider restarts, session pauses, and context switches.
type Workflow struct {
	ID          shared.ID
	Name        string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time // nil until Start() is called
	CompletedAt *time.Time // nil until Complete() or Fail() is called
}

// New validates inputs and returns a Workflow in pending status.
func New(name, description string) (*Workflow, error) {
	if strings.TrimSpace(name) == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "workflow name must not be empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return nil, shared.NewError(shared.CodeInvalidInput, "workflow name must not contain path separators")
	}
	now := time.Now().UTC()
	return &Workflow{
		ID:          shared.NewID(),
		Name:        name,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Start transitions the workflow from pending to running.
func (w *Workflow) Start() error {
	if w.Status != StatusPending {
		return shared.NewError(shared.CodeInvalidInput,
			"only a pending workflow can be started (current status: "+string(w.Status)+")")
	}
	now := time.Now().UTC()
	w.Status = StatusRunning
	w.StartedAt = &now
	w.UpdatedAt = now
	return nil
}

// Complete transitions the workflow from running to completed.
func (w *Workflow) Complete() error {
	if w.Status != StatusRunning {
		return shared.NewError(shared.CodeInvalidInput,
			"only a running workflow can be completed (current status: "+string(w.Status)+")")
	}
	now := time.Now().UTC()
	w.Status = StatusCompleted
	w.CompletedAt = &now
	w.UpdatedAt = now
	return nil
}

// Fail transitions the workflow from running to failed.
func (w *Workflow) Fail() error {
	if w.Status != StatusRunning {
		return shared.NewError(shared.CodeInvalidInput,
			"only a running workflow can be failed (current status: "+string(w.Status)+")")
	}
	now := time.Now().UTC()
	w.Status = StatusFailed
	w.CompletedAt = &now
	w.UpdatedAt = now
	return nil
}

// Pause transitions the workflow from running to paused.
func (w *Workflow) Pause() error {
	if w.Status != StatusRunning {
		return shared.NewError(shared.CodeInvalidInput,
			"only a running workflow can be paused (current status: "+string(w.Status)+")")
	}
	now := time.Now().UTC()
	w.Status = StatusPaused
	w.UpdatedAt = now
	return nil
}

// Resume transitions the workflow from paused to running.
func (w *Workflow) Resume() error {
	if w.Status != StatusPaused {
		return shared.NewError(shared.CodeInvalidInput,
			"only a paused workflow can be resumed (current status: "+string(w.Status)+")")
	}
	now := time.Now().UTC()
	w.Status = StatusRunning
	w.UpdatedAt = now
	return nil
}
