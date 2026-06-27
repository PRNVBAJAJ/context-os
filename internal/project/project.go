package project

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// Project is the core domain object representing a Context OS managed repository.
// It is the aggregate root for all project-level intelligence.
type Project struct {
	ID             shared.ID
	Name           string
	RootPath       string
	Language       string
	RuntimeVersion string
	SchemaVersion  int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// New creates a validated Project. If name is empty it is derived from the
// base of rootPath. rootPath must be an absolute path to an existing directory.
func New(name, rootPath, language string) (*Project, error) {
	if rootPath == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "rootPath is required")
	}
	if !filepath.IsAbs(rootPath) {
		return nil, shared.NewError(shared.CodeInvalidInput, "rootPath must be an absolute path")
	}

	if name == "" {
		name = filepath.Base(rootPath)
	}

	if strings.ContainsAny(name, `/\`) {
		return nil, shared.NewError(shared.CodeInvalidInput, "project name must not contain path separators")
	}
	if name == "" || name == "." {
		return nil, shared.NewError(shared.CodeInvalidInput, "project name is invalid")
	}

	now := time.Now().UTC()
	return &Project{
		ID:             shared.NewID(),
		Name:           name,
		RootPath:       rootPath,
		Language:       language,
		RuntimeVersion: shared.Version,
		SchemaVersion:  0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}
