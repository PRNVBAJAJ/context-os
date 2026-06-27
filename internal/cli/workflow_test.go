package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestWorkflowStart_Success(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-cli-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "start", "implement auth"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("workflow start: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "implement auth") {
		t.Errorf("output should contain workflow name, got: %s", output)
	}
	if !strings.Contains(output, "started") {
		t.Errorf("output should contain 'started', got: %s", output)
	}
}

func TestWorkflowStart_WithDescription(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-desc-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "start", "refactor db", "--description", "Move to UUID keys"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("workflow start --description: %v", err)
	}
}

func TestWorkflowList_Empty(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-list-empty")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("workflow list: %v", err)
	}

	if !strings.Contains(out.String(), "No workflows") {
		t.Errorf("empty list should say 'No workflows', got: %s", out.String())
	}
}

func TestWorkflowList_ShowsStartedWorkflows(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-list-test")
	t.Cleanup(cleanup)

	for _, name := range []string{"implement auth", "refactor db"} {
		start := newRootCommand()
		start.SetOut(&bytes.Buffer{})
		start.SetErr(&bytes.Buffer{})
		start.SetArgs([]string{"workflow", "start", name})
		if err := start.Execute(); err != nil {
			t.Fatalf("workflow start %q: %v", name, err)
		}
	}

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("workflow list: %v", err)
	}

	output := out.String()
	for _, want := range []string{"implement auth", "refactor db", "running", "ID", "NAME", "STATUS"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q:\n%s", want, output)
		}
	}
}

func TestWorkflowStart_RequiresName(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-noargs-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "start"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when name is missing, got nil")
	}
}

func TestWorkflowStart_NotInitialized(t *testing.T) {
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
	cmd.SetArgs([]string{"workflow", "start", "implement auth"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when project not initialized, got nil")
	}
}
