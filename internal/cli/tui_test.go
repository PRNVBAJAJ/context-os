package cli

import (
	"bytes"
	"os"
	"testing"
)

func TestTuiCommand_NotInitialized(t *testing.T) {
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
	cmd.SetArgs([]string{"tui"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when project not initialized, got nil")
	}
}

func TestTuiCommand_RegisteredInRoot(t *testing.T) {
	root := newRootCommand()
	for _, sub := range root.Commands() {
		if sub.Use == "tui" {
			return
		}
	}
	t.Error("'tui' command not registered on root")
}
