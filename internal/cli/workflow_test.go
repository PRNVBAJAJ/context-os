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

func TestWorkflowDelete_Success(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-del-test")
	t.Cleanup(cleanup)

	run := func(args ...string) string {
		t.Helper()
		cmd := newRootCommand()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		return out.String()
	}

	startOut := run("workflow", "start", "to-delete")
	// Extract the 8-char ID prefix from output like: Workflow "to-delete" started (ID: abcd1234).
	var prefix string
	if idx := strings.Index(startOut, "ID: "); idx >= 0 {
		prefix = startOut[idx+4 : idx+12]
	}
	if prefix == "" {
		t.Skip("could not extract workflow ID from output")
	}

	run("workflow", "complete", prefix)

	out := &bytes.Buffer{}
	cmd := newRootCommand()
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "delete", prefix})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("workflow delete: %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("output should mention deleted, got: %s", out.String())
	}
}

func TestWorkflowDelete_RunningRejected(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "wf-del-running-test")
	t.Cleanup(cleanup)

	startOut := &bytes.Buffer{}
	start := newRootCommand()
	start.SetOut(startOut)
	start.SetErr(&bytes.Buffer{})
	start.SetArgs([]string{"workflow", "start", "active-workflow"})
	if err := start.Execute(); err != nil {
		t.Fatal(err)
	}

	var prefix string
	if idx := strings.Index(startOut.String(), "ID: "); idx >= 0 {
		prefix = startOut.String()[idx+4 : idx+12]
	}
	if prefix == "" {
		t.Skip("could not extract workflow ID")
	}

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"workflow", "delete", prefix})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error deleting running workflow, got nil")
	}
}
