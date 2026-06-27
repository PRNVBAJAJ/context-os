package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name     string
		projName string
		rootPath string
		language string
		wantCode shared.Code
	}{
		{
			name:     "valid with explicit name",
			projName: "my-project",
			rootPath: dir,
			language: "go",
		},
		{
			name:     "valid with empty name derives from path",
			projName: "",
			rootPath: dir,
			language: "",
		},
		{
			name:     "empty rootPath",
			projName: "x",
			rootPath: "",
			wantCode: shared.CodeInvalidInput,
		},
		{
			name:     "relative rootPath",
			projName: "x",
			rootPath: "relative/path",
			wantCode: shared.CodeInvalidInput,
		},
		{
			name:     "name with slash",
			projName: "foo/bar",
			rootPath: dir,
			wantCode: shared.CodeInvalidInput,
		},
		{
			name:     "name with backslash",
			projName: `foo\bar`,
			rootPath: dir,
			wantCode: shared.CodeInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := project.New(tc.projName, tc.rootPath, tc.language)

			if tc.wantCode != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var domainErr *shared.Error
				if !errors.As(err, &domainErr) {
					t.Fatalf("expected *shared.Error, got %T: %v", err, err)
				}
				if domainErr.Code != tc.wantCode {
					t.Errorf("Code = %q, want %q", domainErr.Code, tc.wantCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.ID.IsEmpty() {
				t.Error("ID must not be empty")
			}
			if p.Name == "" {
				t.Error("Name must not be empty")
			}
			if p.RootPath != tc.rootPath {
				t.Errorf("RootPath = %q, want %q", p.RootPath, tc.rootPath)
			}
			if p.RuntimeVersion != shared.Version {
				t.Errorf("RuntimeVersion = %q, want %q", p.RuntimeVersion, shared.Version)
			}
			if p.CreatedAt.IsZero() {
				t.Error("CreatedAt must be set")
			}
			if p.SchemaVersion != 0 {
				t.Errorf("SchemaVersion = %d, want 0", p.SchemaVersion)
			}
		})
	}
}

func TestNew_NameDerivedFromPath(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "my-repo")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	p, err := project.New("", dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "my-repo" {
		t.Errorf("Name = %q, want %q", p.Name, "my-repo")
	}
}
