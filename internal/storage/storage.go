package storage

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/project"
)

// Storage is the top-level persistence abstraction for Context OS.
// All access to the SQLite runtime database goes through this interface.
// Obtain an instance via Open() and call Close() when done.
type Storage interface {
	// Projects returns the store for project metadata.
	Projects() ProjectStore
	// Events returns the append-only store for runtime events.
	Events() EventStore
	// Memories returns the store for durable project knowledge.
	Memories() MemoryStore
	// Workflows returns the store for workflow lifecycle records.
	Workflows() WorkflowStore
	// Checkpoints returns the append-only store for recovery snapshots.
	Checkpoints() CheckpointStore
	// FileAccesses returns the store for per-workflow file access tracking.
	FileAccesses() FileAccessStore
	// Close releases the database connection.
	Close() error
}

// EventFilter controls which events are returned by EventStore.List.
type EventFilter struct {
	// Limit caps the number of results. Zero means no limit.
	Limit int
}

// ProjectStore persists and retrieves project metadata.
// It is the only interface with ownership of the project table.
type ProjectStore interface {
	// Create persists a new project record. Returns CodeConflict if a project
	// with the same root_path already exists.
	Create(ctx context.Context, p *project.Project) error
	// GetByPath returns the project whose root_path matches rootPath.
	// Returns CodeNotFound if no matching project exists.
	GetByPath(ctx context.Context, rootPath string) (*project.Project, error)
}
