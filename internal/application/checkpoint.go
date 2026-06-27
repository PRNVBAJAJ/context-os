package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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
	defer func() { _ = store.Close() }()

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
	defer func() { _ = store.Close() }()

	return store.Checkpoints().List(ctx, p.ID, storage.CheckpointFilter{})
}

// RestoreCheckpointOptions carries parameters for the RestoreCheckpoint use case.
type RestoreCheckpointOptions struct {
	RootPath string
	IDPrefix string
}

// RestoreResult holds the checkpoint and its associated workflow, if any.
type RestoreResult struct {
	Checkpoint   *checkpoint.Checkpoint
	WorkflowName string // empty when the checkpoint is project-level
}

// RestoreCheckpoint resolves a checkpoint by ID prefix and returns its state.
// In v0.1, restore surfaces the checkpoint details and workflow context so the
// developer or AI assistant can resume from the recorded note.
func RestoreCheckpoint(ctx context.Context, opts RestoreCheckpointOptions) (*RestoreResult, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	cp, err := resolveCheckpointByPrefix(ctx, store, p.ID, opts.IDPrefix)
	if err != nil {
		return nil, err
	}

	result := &RestoreResult{Checkpoint: cp}

	if !cp.WorkflowID.IsEmpty() {
		w, err := store.Workflows().GetByID(ctx, cp.WorkflowID)
		if err == nil {
			result.WorkflowName = w.Name
		}
	}

	return result, nil
}

// resolveCheckpointByPrefix finds the unique checkpoint whose ID starts with prefix.
func resolveCheckpointByPrefix(ctx context.Context, store storage.Storage, projectID shared.ID, prefix string) (*checkpoint.Checkpoint, error) {
	if prefix == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "checkpoint ID prefix must not be empty")
	}

	all, err := store.Checkpoints().List(ctx, projectID, storage.CheckpointFilter{})
	if err != nil {
		return nil, err
	}

	var matches []*checkpoint.Checkpoint
	for _, cp := range all {
		if strings.HasPrefix(cp.ID.String(), prefix) {
			matches = append(matches, cp)
		}
	}

	switch len(matches) {
	case 0:
		return nil, shared.NewError(shared.CodeNotFound, "no checkpoint found with ID prefix "+prefix)
	case 1:
		return matches[0], nil
	default:
		return nil, shared.NewError(shared.CodeInvalidInput,
			fmt.Sprintf("ambiguous prefix %q matches %d checkpoints — use more characters", prefix, len(matches)))
	}
}
