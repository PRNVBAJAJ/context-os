package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func setupInitializedProject(t *testing.T, name string) (dir string, cleanup func()) {
	t.Helper()
	dir = t.TempDir()

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	init := newRootCommand()
	init.SetOut(&bytes.Buffer{})
	init.SetErr(&bytes.Buffer{})
	init.SetArgs([]string{"init", "--name", name})
	if err := init.Execute(); err != nil {
		os.Chdir(orig) //nolint:errcheck
		t.Fatalf("init: %v", err)
	}

	return dir, func() { os.Chdir(orig) } //nolint:errcheck
}

func TestMemoryAdd_Success(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "mem-cli-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "add", "auth-strategy", "We use JWT with RS256."})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory add: %v", err)
	}

	if !strings.Contains(out.String(), "auth-strategy") {
		t.Errorf("output should contain key name, got: %s", out.String())
	}
}

func TestMemoryAdd_WithTitle(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "title-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "add", "db-schema", "UUID primary keys.", "--title", "DB Schema"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory add --title: %v", err)
	}
}

func TestMemoryList_Empty(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "list-empty-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory list: %v", err)
	}

	if !strings.Contains(out.String(), "No memories") {
		t.Errorf("empty list should say 'No memories', got: %s", out.String())
	}
}

func TestMemoryList_ShowsAddedMemories(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "list-test")
	t.Cleanup(cleanup)

	// Add two memories.
	for _, key := range []string{"auth-strategy", "db-schema"} {
		add := newRootCommand()
		add.SetOut(&bytes.Buffer{})
		add.SetErr(&bytes.Buffer{})
		add.SetArgs([]string{"memory", "add", key, "Content for " + key})
		if err := add.Execute(); err != nil {
			t.Fatalf("memory add %q: %v", key, err)
		}
	}

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory list: %v", err)
	}

	output := out.String()
	for _, want := range []string{"auth-strategy", "db-schema", "KEY", "TITLE"} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q:\n%s", want, output)
		}
	}
}

func TestMemoryAdd_RequiresTwoArgs(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "args-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "add", "only-one-arg"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when content arg is missing, got nil")
	}
}

func TestMemoryAdd_NotInitialized(t *testing.T) {
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
	cmd.SetArgs([]string{"memory", "add", "auth-strategy", "some content"})

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when project not initialized, got nil")
	}
}

func TestMemoryUpdate_Success(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "update-cli-test")
	t.Cleanup(cleanup)

	run := func(args ...string) {
		t.Helper()
		cmd := newRootCommand()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}

	run("memory", "add", "db-driver", "Use mattn/go-sqlite3.")

	out := &bytes.Buffer{}
	cmd := newRootCommand()
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "update", "db-driver", "Use modernc/sqlite — no CGO."})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory update: %v", err)
	}
	if !strings.Contains(out.String(), "updated") {
		t.Errorf("output should mention updated, got: %s", out.String())
	}
}

func TestMemoryUpdate_NotFound(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "update-notfound-test")
	t.Cleanup(cleanup)

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "update", "ghost", "content"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error updating nonexistent key")
	}
}

func TestMemoryDelete_Success(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "del-cli-test")
	t.Cleanup(cleanup)

	run := func(args ...string) {
		t.Helper()
		cmd := newRootCommand()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}

	run("memory", "add", "to-remove", "temporary")

	out := &bytes.Buffer{}
	cmd := newRootCommand()
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "delete", "to-remove"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("memory delete: %v", err)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("output should mention deleted, got: %s", out.String())
	}
}

func TestMemoryList_ShowsPreview(t *testing.T) {
	_, cleanup := setupInitializedProject(t, "preview-test")
	t.Cleanup(cleanup)

	add := newRootCommand()
	add.SetOut(&bytes.Buffer{})
	add.SetErr(&bytes.Buffer{})
	add.SetArgs([]string{"memory", "add", "key1", "This is the content for preview."})
	if err := add.Execute(); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCommand()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"memory", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	output := out.String()
	if !strings.Contains(output, "PREVIEW") {
		t.Errorf("list should have PREVIEW column:\n%s", output)
	}
	if !strings.Contains(output, "This is the content") {
		t.Errorf("list should show content preview:\n%s", output)
	}
}
