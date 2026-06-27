package project

import (
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// Discover walks up the directory tree from startDir, looking for the first
// directory that contains a .context/ runtime directory. It returns the
// absolute path of that directory (the project root).
//
// Returns CodeNotFound if no project root is found before the filesystem root.
func Discover(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", shared.Wrap(shared.CodeInternal, "failed to resolve start directory", err)
	}

	for {
		if IsInitialized(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", shared.NewError(shared.CodeNotFound,
				"no Context OS project found; run 'context init' to initialize one")
		}
		dir = parent
	}
}
