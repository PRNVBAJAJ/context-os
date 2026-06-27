package storage

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/checkpoint"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// CheckpointFilter controls which checkpoints are returned by CheckpointStore.List.
type CheckpointFilter struct {
	// WorkflowID restricts results to a specific workflow. Zero value means all.
	WorkflowID shared.ID
	// Limit caps the number of results. Zero means no limit.
	Limit int
}

// CheckpointStore persists recovery snapshots. Checkpoints are append-only.
type CheckpointStore interface {
	// Create persists a new checkpoint scoped to projectID.
	Create(ctx context.Context, projectID shared.ID, cp *checkpoint.Checkpoint) error
	// List returns checkpoints for the project ordered by created_at descending.
	List(ctx context.Context, projectID shared.ID, filter CheckpointFilter) ([]*checkpoint.Checkpoint, error)
	// GetByID returns the checkpoint with the given ID.
	// Returns CodeNotFound if no match exists.
	GetByID(ctx context.Context, id shared.ID) (*checkpoint.Checkpoint, error)
}
