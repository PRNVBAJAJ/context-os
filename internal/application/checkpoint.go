package application

import (
	"context"
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/checkpoint"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
)

// CreateCheckpointOptions carries parameters for the CreateCheckpoint use case.
type CreateCheckpointOptions struct {
	RootPath         string
	WorkflowIDPrefix string // optional; empty means project-level checkpoint
	Note             string
}

// CreateCheckpoint records a recovery snapshot for the project or a specific workflow.
// If WorkflowIDPrefix is non-empty, it is resolved to a unique workflow by prefix.
func CreateCheckpoint(ctx context.Context, opts CreateCheckpointOptions) (*checkpoint.Checkpoint, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer store.Close()

	workflowID := shared.EmptyID
	if opts.WorkflowIDPrefix != "" {
		w, err := resolveWorkflowByPrefix(ctx, store, p.ID, opts.WorkflowIDPrefix)
		if err != nil {
			return nil, err
		}
		workflowID = w.ID
	}

	cp := checkpoint.New(workflowID, opts.Note)
	if err := store.Checkpoints().Create(ctx, p.ID, cp); err != nil {
		return nil, err
	}

	return cp, nil
}

// ListCheckpointsOptions carries parameters for the ListCheckpoints use case.
type ListCheckpointsOptions struct {
	RootPath string
}

// ListCheckpoints returns all checkpoints for the project, most recent first.
func ListCheckpoints(ctx context.Context, opts ListCheckpointsOptions) ([]*checkpoint.Checkpoint, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer store.Close()

	return store.Checkpoints().List(ctx, p.ID, storage.CheckpointFilter{})
}
