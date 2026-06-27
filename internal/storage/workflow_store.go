package storage

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

// WorkflowFilter controls which workflows are returned by WorkflowStore.List.
type WorkflowFilter struct {
	// Limit caps the number of results. Zero means no limit.
	Limit int
}

// WorkflowStore persists and retrieves workflow records.
// Workflows are scoped to a project via project_id.
type WorkflowStore interface {
	// Create persists a new workflow. Returns CodeConflict on duplicate ID.
	Create(ctx context.Context, projectID shared.ID, w *workflow.Workflow) error
	// Save updates all mutable fields (status, updated_at, started_at, completed_at)
	// for an existing workflow identified by w.ID.
	// Returns CodeNotFound if no workflow with that ID exists.
	Save(ctx context.Context, w *workflow.Workflow) error
	// GetByID returns the workflow with the given ID.
	// Returns CodeNotFound if no match exists.
	GetByID(ctx context.Context, id shared.ID) (*workflow.Workflow, error)
	// List returns workflows for the project ordered by created_at descending
	// (most recent first), optionally capped by filter.Limit.
	List(ctx context.Context, projectID shared.ID, filter WorkflowFilter) ([]*workflow.Workflow, error)
	// Delete removes the workflow with the given ID.
	// Returns CodeNotFound if no workflow with that ID exists.
	Delete(ctx context.Context, id shared.ID) error
}
