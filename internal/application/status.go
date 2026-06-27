package application

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/project"
)

// ProjectStatus is the result of inspecting the current state of a project.
// Additional fields (active workflow, session count, etc.) will be added as
// the corresponding runtime components are implemented.
type ProjectStatus struct {
	Project *project.Project
}

// GetProjectStatus loads the project metadata for the project rooted at rootPath.
func GetProjectStatus(ctx context.Context, rootPath string) (*ProjectStatus, error) {
	p, err := project.Load(rootPath)
	if err != nil {
		return nil, err
	}
	return &ProjectStatus{Project: p}, nil
}
