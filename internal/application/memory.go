package application

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
)

// AddMemoryOptions carries parameters for the AddMemory use case.
type AddMemoryOptions struct {
	RootPath string
	Key      string
	Title    string
	Content  string
}

// AddMemory persists a new memory entry for the project at RootPath.
// It stores the record in SQLite and writes a human-readable markdown file
// to .context/memory/<key>.md so the knowledge is inspectable without Context OS.
func AddMemory(ctx context.Context, opts AddMemoryOptions) (*memory.Memory, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	m, err := memory.New(opts.Key, opts.Title, opts.Content)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	if err := store.Memories().Add(ctx, p.ID, m); err != nil {
		return nil, err
	}

	filePath := filepath.Join(project.Dir(opts.RootPath), "memory", m.Key+".md")
	fileContent := fmt.Sprintf("# %s\n\n%s\n", m.Title, m.Content)
	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to write memory file", err)
	}

	return m, nil
}

// ListMemoriesOptions carries parameters for the ListMemories use case.
type ListMemoriesOptions struct {
	RootPath string
}

// ListMemories returns all memories for the project at RootPath, ordered by creation time.
func ListMemories(ctx context.Context, opts ListMemoriesOptions) ([]*memory.Memory, error) {
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

	return store.Memories().List(ctx, p.ID, storage.MemoryFilter{})
}

// ShowMemoryOptions carries parameters for the ShowMemory use case.
type ShowMemoryOptions struct {
	RootPath string
	Key      string
}

// ShowMemory returns a single memory entry by key.
func ShowMemory(ctx context.Context, opts ShowMemoryOptions) (*memory.Memory, error) {
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

	return store.Memories().GetByKey(ctx, p.ID, opts.Key)
}

// UpdateMemoryOptions carries parameters for the UpdateMemory use case.
type UpdateMemoryOptions struct {
	RootPath string
	Key      string
	Content  string
}

// UpdateMemory replaces the content of an existing memory entry. Both the
// SQLite record and the .context/memory/<key>.md file are updated.
func UpdateMemory(ctx context.Context, opts UpdateMemoryOptions) (*memory.Memory, error) {
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

	if err := store.Memories().Update(ctx, p.ID, opts.Key, opts.Content); err != nil {
		return nil, err
	}

	m, err := store.Memories().GetByKey(ctx, p.ID, opts.Key)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(project.Dir(opts.RootPath), "memory", m.Key+".md")
	fileContent := fmt.Sprintf("# %s\n\n%s\n", m.Title, m.Content)
	if err := os.WriteFile(filePath, []byte(fileContent), 0o644); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to update memory file", err)
	}

	return m, nil
}

// DeleteMemoryOptions carries parameters for the DeleteMemory use case.
type DeleteMemoryOptions struct {
	RootPath string
	Key      string
}

// DeleteMemory removes a memory entry from the database and its markdown file.
func DeleteMemory(ctx context.Context, opts DeleteMemoryOptions) error {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	if err := store.Memories().Delete(ctx, p.ID, opts.Key); err != nil {
		return err
	}

	filePath := filepath.Join(project.Dir(opts.RootPath), "memory", opts.Key+".md")
	if removeErr := os.Remove(filePath); removeErr != nil && !os.IsNotExist(removeErr) {
		return shared.Wrap(shared.CodeInternal, "failed to remove memory file", removeErr)
	}

	return nil
}
