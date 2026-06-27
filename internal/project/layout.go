package project

import (
	"os"
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// ContextDir is the name of the runtime directory created inside every project root.
const ContextDir = ".context"

// contextSubdirs is the canonical list of directories created by CreateLayout.
// Each is owned by exactly one runtime service (Chapter 9).
var contextSubdirs = []string{
	"workflows",
	"sessions",
	"memory",
	"artifacts",
	"checkpoints",
	"providers",
	"events",
	"logs",
	"cache",
	"temp",
	"plugins",
}

// Dir returns the absolute path to the .context/ directory for rootPath.
func Dir(rootPath string) string {
	return filepath.Join(rootPath, ContextDir)
}

// IsInitialized reports whether rootPath already contains a .context/ directory.
func IsInitialized(rootPath string) bool {
	info, err := os.Stat(Dir(rootPath))
	return err == nil && info.IsDir()
}

// CreateLayout creates the .context/ directory structure inside rootPath.
// It returns CodeConflict if the .context/ directory already exists.
func CreateLayout(rootPath string) error {
	contextDir := Dir(rootPath)

	if IsInitialized(rootPath) {
		return shared.NewError(shared.CodeConflict, "project is already initialized")
	}

	if err := os.MkdirAll(contextDir, 0o755); err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to create .context directory", err)
	}

	for _, sub := range contextSubdirs {
		path := filepath.Join(contextDir, sub)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return shared.Wrap(shared.CodeInternal, "failed to create subdirectory "+sub, err)
		}
	}

	return nil
}
