package application

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
)

// InitOptions carries parameters for the InitializeProject use case.
type InitOptions struct {
	// Name is the human-readable project name.
	// If empty, it is derived from the base of RootPath.
	Name string
	// RootPath is the absolute path to the project root directory.
	RootPath string
	// Language is the primary programming language (e.g. "go", "python").
	// Optional; an empty string is valid.
	Language string
}

// InitializeProject initializes a new Context OS project at opts.RootPath.
//
// Sequence:
//  1. Validate opts and construct the Project domain object.
//  2. Guard against double-initialization.
//  3. Create the .context/ directory layout.
//  4. Write project.yaml.
//  5. Open SQLite at .context/runtime.db and run schema migrations.
//  6. Persist the project record.
func InitializeProject(ctx context.Context, opts InitOptions) (*project.Project, error) {
	p, err := project.New(opts.Name, opts.RootPath, opts.Language)
	if err != nil {
		return nil, err
	}

	if project.IsInitialized(opts.RootPath) {
		return nil, shared.NewError(shared.CodeConflict, "project is already initialized")
	}

	if err := project.CreateLayout(opts.RootPath); err != nil {
		return nil, err
	}

	if err := project.Save(p); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	if err := store.Projects().Create(ctx, p); err != nil {
		return nil, err
	}

	payload := fmt.Sprintf(`{"project_id":%q,"project_name":%q}`, p.ID.String(), p.Name)
	e := event.New(event.TypeProjectInitialized, payload)
	if err := store.Events().Append(ctx, e); err != nil {
		return nil, err
	}

	return p, nil
}
