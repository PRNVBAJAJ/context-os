package application_test

import (
	"context"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/event"
)

func TestRunDoctor_OnFreshProject(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	if _, err := application.InitializeProject(ctx, application.InitOptions{
		Name:     "doctor-test",
		RootPath: dir,
		Language: "go",
	}); err != nil {
		t.Fatalf("InitializeProject: %v", err)
	}

	result, err := application.RunDoctor(ctx, dir)
	if err != nil {
		t.Fatalf("RunDoctor: %v", err)
	}

	if result.ProjectName != "doctor-test" {
		t.Errorf("ProjectName = %q, want %q", result.ProjectName, "doctor-test")
	}
	if !result.DatabaseOK {
		t.Error("DatabaseOK should be true")
	}
	if result.EventCount != 1 {
		t.Errorf("EventCount = %d, want 1 (project.initialized)", result.EventCount)
	}
	if len(result.RecentEvents) != 1 {
		t.Fatalf("RecentEvents len = %d, want 1", len(result.RecentEvents))
	}
	if result.RecentEvents[0].Type != event.TypeProjectInitialized {
		t.Errorf("first event type = %q, want %q",
			result.RecentEvents[0].Type, event.TypeProjectInitialized)
	}
}

func TestRunDoctor_NotInitialized(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	_, err := application.RunDoctor(ctx, dir)
	if err == nil {
		t.Error("expected error for uninitialized project, got nil")
	}
}
