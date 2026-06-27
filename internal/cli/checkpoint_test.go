package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestCheckpointCreate_ProjectLevel(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "cp-cli-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"checkpoint", "create", "--note", "before refactor"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("checkpoint create: %v", err)
	}

	if !strings.Contains(out.String(), "Checkpoint created") {
		t.Errorf("output should confirm creation, got: %s", out.String())
	}
}

func TestCheckpointCreate_WithWorkflow(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "cp-wf-cli")
	t.Cleanup(cleanup)

	// Start a workflow first.
	start := newRootCommand()
	startOut := &bytes.Buffer{}
	start.SetOut(startOut)
	start.SetErr(&bytes.Buffer{})
	start.SetArgs([]string{"workflow", "start", "implement auth"})
	if err := start.Execute(); err != nil {
		t.Fatalf("workflow start: %v", err)
	}

	// Extract ID prefix from "Workflow ... started (ID: abcd1234)."
	output := startOut.String()
	idStart := strings.Index(output, "ID: ") + 4
	idPrefix := output[idStart : idStart+8]

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"checkpoint", "create", "--workflow", idPrefix, "--note", "mid-auth"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("checkpoint create --workflow: %v", err)
	}

	if !strings.Contains(out.String(), "Checkpoint created") {
		t.Errorf("output should confirm creation, got: %s", out.String())
	}
}

func TestCheckpointList_Empty(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "cp-list-empty")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"checkpoint", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("checkpoint list: %v", err)
	}

	if !strings.Contains(out.String(), "No checkpoints") {
		t.Errorf("empty list should say 'No checkpoints', got: %s", out.String())
	}
}

func TestCheckpointList_ShowsCheckpoints(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "cp-list-test")
	t.Cleanup(cleanup)

	for _, note := range []string{"first snapshot", "second snapshot"} {
		add := newRootCommand()
		add.SetOut(&bytes.Buffer{})
		add.SetErr(&bytes.Buffer{})
		add.SetArgs([]string{"checkpoint", "create", "--note", note})
		if err := add.Execute(); err != nil {
			t.Fatalf("checkpoint create: %v", err)
		}
	}

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"checkpoint", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("checkpoint list: %v", err)
	}

	output := out.String()
	for _, want := range []string{"first snapshot", "second snapshot", "ID", "NOTE", "CREATED"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q:\n%s", want, output)
		}
	}
}

func TestWorkflowTransitions_CLI(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "trans-cli-test")
	t.Cleanup(cleanup)

	// Start.
	start := newRootCommand()
	startOut := &bytes.Buffer{}
	start.SetOut(startOut)
	start.SetErr(&bytes.Buffer{})
	start.SetArgs([]string{"workflow", "start", "implement auth"})
	if err := start.Execute(); err != nil {
		t.Fatalf("start: %v", err)
	}

	// Extract ID prefix.
	output := startOut.String()
	idStart := strings.Index(output, "ID: ") + 4
	idPrefix := output[idStart : idStart+8]

	// Pause.
	pause := newRootCommand()
	pause.SetOut(&bytes.Buffer{})
	pause.SetErr(&bytes.Buffer{})
	pause.SetArgs([]string{"workflow", "pause", idPrefix})
	if err := pause.Execute(); err != nil {
		t.Fatalf("pause: %v", err)
	}

	// Resume.
	resume := newRootCommand()
	resume.SetOut(&bytes.Buffer{})
	resume.SetErr(&bytes.Buffer{})
	resume.SetArgs([]string{"workflow", "resume", idPrefix})
	if err := resume.Execute(); err != nil {
		t.Fatalf("resume: %v", err)
	}

	// Complete.
	complete := newRootCommand()
	completeOut := &bytes.Buffer{}
	complete.SetOut(completeOut)
	complete.SetErr(&bytes.Buffer{})
	complete.SetArgs([]string{"workflow", "complete", idPrefix})
	if err := complete.Execute(); err != nil {
		t.Fatalf("complete: %v", err)
	}

	if !strings.Contains(completeOut.String(), "completed") {
		t.Errorf("complete output should say 'completed', got: %s", completeOut.String())
	}

	// List — should show completed.
	list := newRootCommand()
	listOut := &bytes.Buffer{}
	list.SetOut(listOut)
	list.SetErr(&bytes.Buffer{})
	list.SetArgs([]string{"workflow", "list"})
	if err := list.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(listOut.String(), "completed") {
		t.Errorf("list should show completed status, got: %s", listOut.String())
	}
}
