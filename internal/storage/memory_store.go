package storage

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// MemoryFilter controls which memories are returned by MemoryStore.List.
type MemoryFilter struct {
	// Limit caps the number of results. Zero means no limit.
	Limit int
}

// MemoryStore persists and retrieves durable project knowledge entries.
// Keys are unique within a project — (project_id, key) is a UNIQUE constraint.
type MemoryStore interface {
	// Add persists a new memory scoped to projectID.
	// Returns CodeConflict if the key already exists for that project.
	Add(ctx context.Context, projectID shared.ID, m *memory.Memory) error
	// List returns all memories for the project ordered by created_at ascending.
	List(ctx context.Context, projectID shared.ID, filter MemoryFilter) ([]*memory.Memory, error)
	// GetByKey returns the memory matching key for the project.
	// Returns CodeNotFound if no match exists.
	GetByKey(ctx context.Context, projectID shared.ID, key string) (*memory.Memory, error)
	// Update replaces the content of an existing memory identified by key.
	// Returns CodeNotFound if no match exists.
	Update(ctx context.Context, projectID shared.ID, key string, content string) error
	// Delete removes the memory identified by key from the project.
	// Returns CodeNotFound if no match exists.
	Delete(ctx context.Context, projectID shared.ID, key string) error
}
