package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDoctorCommand_OnInitializedProject(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	// Initialize first.
	init := newRootCommand()
	init.SetOut(&bytes.Buffer{})
	init.SetErr(&bytes.Buffer{})
	init.SetArgs([]string{"init", "--name", "doc-project"})
	if err := init.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Run doctor.
	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"doctor"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor: %v", err)
	}

	output := out.String()
	for _, want := range []string{"OK", "doc-project", "Events:", "project.initialized"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q:\n%s", want, output)
		}
	}
}

func TestDoctorCommand_NotInitialized(t *testing.T) {
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
	cmd.SetArgs([]string{"doctor"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when project not initialized, got nil")
	}
}
