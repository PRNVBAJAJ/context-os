package project

import (
	"os"
	"path/filepath"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"gopkg.in/yaml.v3"
)

const projectFileName = "project.yaml"

// projectFile is the YAML-serializable representation of a Project.
// RootPath is intentionally excluded — it is derived from the file's location,
// keeping project.yaml portable across machines.
type projectFile struct {
	ID             string    `yaml:"id"`
	Name           string    `yaml:"name"`
	Language       string    `yaml:"language"`
	RuntimeVersion string    `yaml:"runtime_version"`
	SchemaVersion  int       `yaml:"schema_version"`
	CreatedAt      time.Time `yaml:"created_at"`
	UpdatedAt      time.Time `yaml:"updated_at"`
}

// Save writes p to .context/project.yaml inside p.RootPath.
// The .context/ directory must already exist.
func Save(p *Project) error {
	f := projectFile{
		ID:             p.ID.String(),
		Name:           p.Name,
		Language:       p.Language,
		RuntimeVersion: p.RuntimeVersion,
		SchemaVersion:  p.SchemaVersion,
		CreatedAt:      p.CreatedAt.UTC(),
		UpdatedAt:      p.UpdatedAt.UTC(),
	}

	data, err := yaml.Marshal(f)
	if err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to marshal project.yaml", err)
	}

	path := filepath.Join(Dir(p.RootPath), projectFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to write project.yaml", err)
	}

	return nil
}

// Load reads .context/project.yaml from rootPath and returns the Project.
// rootPath must be an absolute path to the project root (not the .context/ dir).
func Load(rootPath string) (*Project, error) {
	path := filepath.Join(Dir(rootPath), projectFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, shared.NewError(shared.CodeNotFound, "project.yaml not found; run 'context init' first")
		}
		return nil, shared.Wrap(shared.CodeInternal, "failed to read project.yaml", err)
	}

	var f projectFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse project.yaml", err)
	}

	if f.ID == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "project.yaml is missing required field: id")
	}

	return &Project{
		ID:             shared.ID(f.ID),
		Name:           f.Name,
		RootPath:       rootPath,
		Language:       f.Language,
		RuntimeVersion: f.RuntimeVersion,
		SchemaVersion:  f.SchemaVersion,
		CreatedAt:      f.CreatedAt.UTC(),
		UpdatedAt:      f.UpdatedAt.UTC(),
	}, nil
}
