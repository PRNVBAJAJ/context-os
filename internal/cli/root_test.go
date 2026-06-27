package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestRootCommand_Metadata(t *testing.T) {
	cmd := newRootCommand()

	if cmd.Use != "context" {
		t.Errorf("Use = %q, want %q", cmd.Use, "context")
	}
	if cmd.Version != shared.Version {
		t.Errorf("Version = %q, want %q", cmd.Version, shared.Version)
	}
	if cmd.Short == "" {
		t.Error("Short description must not be empty")
	}
	if !cmd.SilenceUsage {
		t.Error("SilenceUsage must be true to avoid cluttered error output")
	}
}

func TestRootCommand_HelpContainsUse(t *testing.T) {
	cmd := newRootCommand()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	// --help prints usage and returns nil; cobra does not exit in this path.
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(--help) returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Context OS") {
		t.Errorf("help output should contain %q, got:\n%s", "Context OS", output)
	}
}

func TestRootCommand_UnknownFlagReturnsError(t *testing.T) {
	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--unknown-flag"})

	if err := cmd.Execute(); err == nil {
		t.Error("Execute with unknown flag should return an error")
	}
}
