package storage

import (
	"context"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// FileAccess records how many times a file was touched during a workflow.
type FileAccess struct {
	WorkflowID  shared.ID
	Filepath    string
	AccessCount int
	LastAccess  time.Time
}

// FileAccessStore tracks which files are touched during a workflow.
type FileAccessStore interface {
	// Record upserts a file access entry, incrementing the counter by one.
	Record(ctx context.Context, workflowID shared.ID, filepath string) error
	// HotFiles returns the most-accessed files for a workflow, ordered by
	// access_count descending. limit=0 means no cap.
	HotFiles(ctx context.Context, workflowID shared.ID, limit int) ([]FileAccess, error)
}
