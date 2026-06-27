package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand_SuccessInTempDir(t *testing.T) {
	dir := t.TempDir()

	// Change to the temp dir so os.Getwd() resolves correctly inside RunE.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"init"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(out.String(), "Initialized") {
		t.Errorf("expected success message, got: %s", out.String())
	}

	// Verify .context/ was actually created.
	if _, err := os.Stat(filepath.Join(dir, ".context")); err != nil {
		t.Errorf(".context/ should exist after init: %v", err)
	}
}

func TestInitCommand_ConflictOnDoubleInit(t *testing.T) {
	dir := t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	run := func() error {
		cmd := newRootCommand()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"init"})
		return cmd.Execute()
	}

	if err := run(); err != nil {
		t.Fatalf("first init: %v", err)
	}

	if err := run(); err == nil {
		t.Error("expected error on second init, got nil")
	}
}

func TestInitCommand_WithFlags(t *testing.T) {
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
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"init", "--name", "my-project", "--language", "go"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !strings.Contains(out.String(), "my-project") {
		t.Errorf("success message should mention project name, got: %s", out.String())
	}
}
