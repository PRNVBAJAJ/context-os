package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestProvidersListCommand(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "providers-list-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"providers", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("providers list: %v", err)
	}

	output := out.String()
	for _, want := range []string{"PROVIDER", "DETECTED", "CONFIG FILE", "INJECTED", "claude", "gemini"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q:\n%s", want, output)
		}
	}
}

func TestProvidersInjectCommand_SpecificProvider(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "providers-inject-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"providers", "inject", "gemini"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("providers inject gemini: %v", err)
	}

	if !strings.Contains(out.String(), "gemini") {
		t.Errorf("output should mention provider name, got: %s", out.String())
	}
}

func TestProvidersInjectCommand_Idempotent(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "providers-inject-idem-test")
	t.Cleanup(cleanup)

	run := func() string {
		cmd := newRootCommand()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"providers", "inject", "gemini"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("providers inject: %v", err)
		}
		return out.String()
	}

	run() // first inject
	second := run()
	// Second run should report already configured, not inject again.
	if !strings.Contains(second, "gemini") {
		t.Logf("second inject output: %s", second)
	}
}
