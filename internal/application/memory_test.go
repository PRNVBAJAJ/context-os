package application_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestAddMemory_Success(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "mem-test",
		RootPath: dir,
	}); err != nil {
		t.Fatalf("InitializeProject: %v", err)
	}

	m, err := application.AddMemory(ctx, application.AddMemoryOptions{
		RootPath: dir,
		Key:      "auth-strategy",
		Title:    "Auth Strategy",
		Content:  "We use JWT with RS256 signed by a rotating key pair.",
	})
	if err != nil {
		t.Fatalf("AddMemory: %v", err)
	}
	if m.Key != "auth-strategy" {
		t.Errorf("Key = %q, want %q", m.Key, "auth-strategy")
	}
	if m.ID.IsEmpty() {
		t.Error("ID should not be empty")
	}
}

func TestAddMemory_WritesMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "file-test",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := application.AddMemory(ctx, application.AddMemoryOptions{
		RootPath: dir,
		Key:      "db-schema",
		Content:  "UUID primary keys on all tables.",
	}); err != nil {
		t.Fatalf("AddMemory: %v", err)
	}

	filePath := filepath.Join(project.Dir(dir), "memory", "db-schema.md")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("memory file should exist at %s: %v", filePath, err)
	}
	if len(data) == 0 {
		t.Error("memory file should not be empty")
	}
}

func TestAddMemory_ConflictOnDuplicateKey(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "dup-test",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	opts := application.AddMemoryOptions{
		RootPath: dir,
		Key:      "auth-strategy",
		Content:  "First version.",
	}
	if _, err := application.AddMemory(ctx, opts); err != nil {
		t.Fatalf("first AddMemory: %v", err)
	}

	opts.Content = "Second version."
	_, err := application.AddMemory(ctx, opts)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeConflict {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeConflict)
	}
}

func TestListMemories_ReturnsAddedMemories(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "list-test",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	keys := []string{"auth-strategy", "db-schema", "api-conventions"}
	for _, key := range keys {
		if _, err := application.AddMemory(ctx, application.AddMemoryOptions{
			RootPath: dir,
			Key:      key,
			Content:  "Content for " + key,
		}); err != nil {
			t.Fatalf("AddMemory(%q): %v", key, err)
		}
	}

	memories, err := application.ListMemories(ctx, application.ListMemoriesOptions{RootPath: dir})
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if len(memories) != 3 {
		t.Errorf("ListMemories returned %d memories, want 3", len(memories))
	}
}

func TestListMemories_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "empty-test",
		RootPath: dir,
	}); err != nil {
		t.Fatal(err)
	}

	memories, err := application.ListMemories(ctx, application.ListMemoriesOptions{RootPath: dir})
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if len(memories) != 0 {
		t.Errorf("expected 0 memories, got %d", len(memories))
	}
}
