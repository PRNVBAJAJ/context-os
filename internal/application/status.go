package application

import (
	"context"
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/checkpoint"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

// ProjectStatus is the result of inspecting the current state of a project.
type ProjectStatus struct {
	Project        *project.Project
	ActiveWorkflow *workflow.Workflow
	MemoryCount    int
	LastCheckpoint *checkpoint.Checkpoint
	HotFiles       []HotFile // top files touched in the active workflow
}

// GetProjectStatus loads the project metadata and runtime state for rootPath.
func GetProjectStatus(ctx context.Context, rootPath string) (*ProjectStatus, error) {
	p, err := project.Load(rootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(rootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	status := &ProjectStatus{Project: p}

	// Active workflow: first running one found.
	workflows, err := store.Workflows().List(ctx, p.ID, storage.WorkflowFilter{})
	if err != nil {
		return nil, err
	}
	for _, w := range workflows {
		if w.Status == workflow.StatusRunning {
			status.ActiveWorkflow = w
			break
		}
	}

	// Memory count.
	memories, err := store.Memories().List(ctx, p.ID, storage.MemoryFilter{})
	if err != nil {
		return nil, err
	}
	status.MemoryCount = len(memories)

	// Last checkpoint.
	checkpoints, err := store.Checkpoints().List(ctx, p.ID, storage.CheckpointFilter{})
	if err != nil {
		return nil, err
	}
	if len(checkpoints) > 0 {
		status.LastCheckpoint = checkpoints[len(checkpoints)-1]
	}

	// Hot files for the active workflow (top 5).
	if status.ActiveWorkflow != nil {
		raw, _ := store.FileAccesses().HotFiles(ctx, status.ActiveWorkflow.ID, 5)
		for _, r := range raw {
			status.HotFiles = append(status.HotFiles, HotFile{
				Filepath:    r.Filepath,
				AccessCount: r.AccessCount,
			})
		}
	}

	return status, nil
}
