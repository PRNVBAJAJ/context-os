package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestStatusCommand_ShowsProjectInfo(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	// Initialize the project first.
	init := newRootCommand()
	init.SetOut(&bytes.Buffer{})
	init.SetErr(&bytes.Buffer{})
	init.SetArgs([]string{"init", "--name", "status-project", "--language", "go"})
	if err := init.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Now run status.
	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status: %v", err)
	}

	output := out.String()
	for _, want := range []string{"status-project", "go", "0.1.0-dev"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q:\n%s", want, output)
		}
	}
}

func TestStatusCommand_WorksFromSubdirectory(t *testing.T) {
	root := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	// Init from root.
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	init := newRootCommand()
	init.SetOut(&bytes.Buffer{})
	init.SetErr(&bytes.Buffer{})
	init.SetArgs([]string{"init", "--name", "sub-test"})
	if err := init.Execute(); err != nil {
		t.Fatal(err)
	}

	// Create a nested subdirectory and run status from there.
	sub := root + "/internal/pkg"
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sub); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status from subdir: %v", err)
	}

	if !strings.Contains(out.String(), "sub-test") {
		t.Errorf("expected project name in output:\n%s", out.String())
	}
}

func TestStatusCommand_NotInitialized(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when project not initialized, got nil")
	}
}
