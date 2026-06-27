package storage_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/checkpoint"
	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

func openTestDB(t *testing.T) storage.Storage {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), filepath.Join(dir, "runtime.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func makeProject(t *testing.T, rootPath string) *project.Project {
	t.Helper()
	p, err := project.New("test-project", rootPath, "go")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestOpen_CreatesDatabase(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		t.Fatal("Open returned nil storage")
	}
}

func TestOpen_RunsMigrations(t *testing.T) {
	// Opening twice should not fail — migrations are idempotent.
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime.db")

	db1, err := storage.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	db1.Close()

	db2, err := storage.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("second Open (idempotent): %v", err)
	}
	db2.Close()
}

func TestProjectStore_CreateAndGetByPath(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	dir := t.TempDir()
	p := makeProject(t, dir)
	p.Language = "go"

	if err := db.Projects().Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := db.Projects().GetByPath(ctx, dir)
	if err != nil {
		t.Fatalf("GetByPath: %v", err)
	}

	if got.ID != p.ID {
		t.Errorf("ID = %q, want %q", got.ID, p.ID)
	}
	if got.Name != p.Name {
		t.Errorf("Name = %q, want %q", got.Name, p.Name)
	}
	if got.RootPath != dir {
		t.Errorf("RootPath = %q, want %q", got.RootPath, dir)
	}
	if got.Language != "go" {
		t.Errorf("Language = %q, want %q", got.Language, "go")
	}
	if got.RuntimeVersion != shared.Version {
		t.Errorf("RuntimeVersion = %q, want %q", got.RuntimeVersion, shared.Version)
	}
}

func TestProjectStore_Create_ConflictOnDuplicatePath(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	dir := t.TempDir()
	p1 := makeProject(t, dir)
	p2 := makeProject(t, dir) // same rootPath, different ID

	if err := db.Projects().Create(ctx, p1); err != nil {
		t.Fatalf("Create p1: %v", err)
	}

	err := db.Projects().Create(ctx, p2)
	if err == nil {
		t.Fatal("expected error on duplicate root_path, got nil")
	}

	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeConflict {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeConflict)
	}
}

func TestProjectStore_GetByPath_NotFound(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.Projects().GetByPath(ctx, "/nonexistent/path")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeNotFound {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeNotFound)
	}
}

func TestProjectStore_TimestampsRoundTrip(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	dir := t.TempDir()
	p := makeProject(t, dir)
	// Pin to a specific time to verify round-trip precision.
	p.CreatedAt = time.Date(2026, 6, 27, 10, 30, 0, 0, time.UTC)
	p.UpdatedAt = time.Date(2026, 6, 27, 11, 45, 0, 0, time.UTC)

	if err := db.Projects().Create(ctx, p); err != nil {
		t.Fatal(err)
	}

	got, err := db.Projects().GetByPath(ctx, dir)
	if err != nil {
		t.Fatal(err)
	}

	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, p.CreatedAt)
	}
	if !got.UpdatedAt.Equal(p.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, p.UpdatedAt)
	}
}

// EventStore tests

func TestEventStore_AppendAndList(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	e := event.New(event.TypeProjectInitialized, `{"project_id":"test-123"}`)

	if err := db.Events().Append(ctx, e); err != nil {
		t.Fatalf("Append: %v", err)
	}

	events, err := db.Events().List(ctx, storage.EventFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("List() returned %d events, want 1", len(events))
	}

	got := events[0]
	if got.ID != e.ID {
		t.Errorf("ID = %q, want %q", got.ID, e.ID)
	}
	if got.Type != event.TypeProjectInitialized {
		t.Errorf("Type = %q, want %q", got.Type, event.TypeProjectInitialized)
	}
	if got.Payload != e.Payload {
		t.Errorf("Payload = %q, want %q", got.Payload, e.Payload)
	}
	if !got.Timestamp.Equal(e.Timestamp.Truncate(time.Second)) {
		// RFC3339 truncates to seconds.
		t.Errorf("Timestamp = %v, want ~%v", got.Timestamp, e.Timestamp)
	}
	if !got.WorkflowID.IsEmpty() {
		t.Errorf("WorkflowID should be empty for project-level events, got %q", got.WorkflowID)
	}
}

func TestEventStore_ListEmpty(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	events, err := db.Events().List(ctx, storage.EventFilter{})
	if err != nil {
		t.Fatalf("List on empty store: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestEventStore_ListWithLimit(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	for range 5 {
		e := event.New(event.TypeProjectInitialized, "{}")
		if err := db.Events().Append(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	events, err := db.Events().List(ctx, storage.EventFilter{Limit: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Errorf("List(Limit=3) returned %d events, want 3", len(events))
	}
}

// MemoryStore tests

func makeMemory(t *testing.T, key, content string) *memory.Memory {
	t.Helper()
	m, err := memory.New(key, "", content)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMemoryStore_AddAndList(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	m := makeMemory(t, "auth-strategy", "We use JWT with RS256.")
	if err := db.Memories().Add(ctx, projectID, m); err != nil {
		t.Fatalf("Add: %v", err)
	}

	memories, err := db.Memories().List(ctx, projectID, storage.MemoryFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(memories) != 1 {
		t.Fatalf("List returned %d memories, want 1", len(memories))
	}

	got := memories[0]
	if got.Key != "auth-strategy" {
		t.Errorf("Key = %q, want %q", got.Key, "auth-strategy")
	}
	if got.Title != "auth-strategy" {
		t.Errorf("Title = %q, want %q", got.Title, "auth-strategy")
	}
	if got.Content != "We use JWT with RS256." {
		t.Errorf("Content = %q, want %q", got.Content, "We use JWT with RS256.")
	}
}

func TestMemoryStore_Add_ConflictOnDuplicateKey(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	m1 := makeMemory(t, "auth-strategy", "First version.")
	m2 := makeMemory(t, "auth-strategy", "Second version.")

	if err := db.Memories().Add(ctx, projectID, m1); err != nil {
		t.Fatalf("Add m1: %v", err)
	}

	err := db.Memories().Add(ctx, projectID, m2)
	if err == nil {
		t.Fatal("expected error on duplicate key, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeConflict {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeConflict)
	}
}

func TestMemoryStore_GetByKey(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	m := makeMemory(t, "db-schema", "UUID primary keys everywhere.")
	if err := db.Memories().Add(ctx, projectID, m); err != nil {
		t.Fatal(err)
	}

	got, err := db.Memories().GetByKey(ctx, projectID, "db-schema")
	if err != nil {
		t.Fatalf("GetByKey: %v", err)
	}
	if got.ID != m.ID {
		t.Errorf("ID = %q, want %q", got.ID, m.ID)
	}
}

func TestMemoryStore_GetByKey_NotFound(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	_, err := db.Memories().GetByKey(ctx, projectID, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeNotFound {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeNotFound)
	}
}

func TestMemoryStore_ScopedToProject(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectA := shared.NewID()
	projectB := shared.NewID()

	m := makeMemory(t, "auth-strategy", "Content for A.")
	if err := db.Memories().Add(ctx, projectA, m); err != nil {
		t.Fatal(err)
	}

	// Project B should have no memories.
	memories, err := db.Memories().List(ctx, projectB, storage.MemoryFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(memories) != 0 {
		t.Errorf("expected 0 memories for project B, got %d", len(memories))
	}
}

func TestMemoryStore_ListEmpty(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	memories, err := db.Memories().List(ctx, projectID, storage.MemoryFilter{})
	if err != nil {
		t.Fatalf("List on empty store: %v", err)
	}
	if len(memories) != 0 {
		t.Errorf("expected 0 memories, got %d", len(memories))
	}
}

// WorkflowStore tests

func makeWorkflow(t *testing.T, name string) *workflow.Workflow {
	t.Helper()
	w, err := workflow.New(name, "")
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestWorkflowStore_CreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	w := makeWorkflow(t, "implement auth")
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	if err := db.Workflows().Create(ctx, projectID, w); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := db.Workflows().GetByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != w.ID {
		t.Errorf("ID = %q, want %q", got.ID, w.ID)
	}
	if got.Name != "implement auth" {
		t.Errorf("Name = %q, want %q", got.Name, "implement auth")
	}
	if got.Status != workflow.StatusRunning {
		t.Errorf("Status = %q, want %q", got.Status, workflow.StatusRunning)
	}
	if got.StartedAt == nil {
		t.Error("StartedAt should not be nil after Start()")
	}
}

func TestWorkflowStore_GetByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	_, err := db.Workflows().GetByID(ctx, shared.NewID())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeNotFound {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeNotFound)
	}
}

func TestWorkflowStore_List(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	names := []string{"implement auth", "refactor db", "add tests"}
	for _, name := range names {
		w := makeWorkflow(t, name)
		if err := w.Start(); err != nil {
			t.Fatal(err)
		}
		if err := db.Workflows().Create(ctx, projectID, w); err != nil {
			t.Fatalf("Create(%q): %v", name, err)
		}
	}

	workflows, err := db.Workflows().List(ctx, projectID, storage.WorkflowFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(workflows) != 3 {
		t.Errorf("List returned %d workflows, want 3", len(workflows))
	}
}

func TestWorkflowStore_ListEmpty(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	workflows, err := db.Workflows().List(ctx, projectID, storage.WorkflowFilter{})
	if err != nil {
		t.Fatalf("List on empty store: %v", err)
	}
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows, got %d", len(workflows))
	}
}

func TestWorkflowStore_ScopedToProject(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	projectA := shared.NewID()
	projectB := shared.NewID()

	w := makeWorkflow(t, "implement auth")
	w.Start() //nolint:errcheck
	if err := db.Workflows().Create(ctx, projectA, w); err != nil {
		t.Fatal(err)
	}

	workflows, err := db.Workflows().List(ctx, projectB, storage.WorkflowFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows for project B, got %d", len(workflows))
	}
}

func TestWorkflowStore_NullableTimestamps(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	// Pending workflow has no StartedAt or CompletedAt.
	w := makeWorkflow(t, "pending workflow")
	if err := db.Workflows().Create(ctx, projectID, w); err != nil {
		t.Fatal(err)
	}

	got, err := db.Workflows().GetByID(ctx, w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.StartedAt != nil {
		t.Errorf("StartedAt should be nil for pending workflow, got %v", got.StartedAt)
	}
	if got.CompletedAt != nil {
		t.Errorf("CompletedAt should be nil for pending workflow, got %v", got.CompletedAt)
	}
}

func TestWorkflowStore_Save_UpdatesStatus(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	w := makeWorkflow(t, "implement auth")
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	if err := db.Workflows().Create(ctx, projectID, w); err != nil {
		t.Fatal(err)
	}

	if err := w.Complete(); err != nil {
		t.Fatal(err)
	}
	if err := db.Workflows().Save(ctx, w); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := db.Workflows().GetByID(ctx, w.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != workflow.StatusCompleted {
		t.Errorf("Status = %q, want %q", got.Status, workflow.StatusCompleted)
	}
	if got.CompletedAt == nil {
		t.Error("CompletedAt should be set after Complete()+Save()")
	}
}

func TestWorkflowStore_Save_NotFound(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	w := makeWorkflow(t, "ghost workflow")
	err := db.Workflows().Save(ctx, w)
	if err == nil {
		t.Fatal("expected error saving nonexistent workflow, got nil")
	}
	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if domainErr.Code != shared.CodeNotFound {
		t.Errorf("Code = %q, want %q", domainErr.Code, shared.CodeNotFound)
	}
}

// CheckpointStore tests

func TestCheckpointStore_CreateAndList(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()
	wfID := shared.NewID()

	cp := checkpoint.New(wfID, "before database refactor")
	if err := db.Checkpoints().Create(ctx, projectID, cp); err != nil {
		t.Fatalf("Create: %v", err)
	}

	checkpoints, err := db.Checkpoints().List(ctx, projectID, storage.CheckpointFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("List returned %d checkpoints, want 1", len(checkpoints))
	}

	got := checkpoints[0]
	if got.ID != cp.ID {
		t.Errorf("ID = %q, want %q", got.ID, cp.ID)
	}
	if got.WorkflowID != wfID {
		t.Errorf("WorkflowID = %q, want %q", got.WorkflowID, wfID)
	}
	if got.Note != "before database refactor" {
		t.Errorf("Note = %q, want %q", got.Note, "before database refactor")
	}
}

func TestCheckpointStore_ProjectLevel(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()

	cp := checkpoint.New(shared.EmptyID, "project-level snapshot")
	if err := db.Checkpoints().Create(ctx, projectID, cp); err != nil {
		t.Fatalf("Create: %v", err)
	}

	checkpoints, err := db.Checkpoints().List(ctx, projectID, storage.CheckpointFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(checkpoints))
	}
	if !checkpoints[0].WorkflowID.IsEmpty() {
		t.Errorf("WorkflowID should be empty for project-level checkpoint, got %q", checkpoints[0].WorkflowID)
	}
}

func TestCheckpointStore_FilterByWorkflow(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectID := shared.NewID()
	wfA := shared.NewID()
	wfB := shared.NewID()

	if err := db.Checkpoints().Create(ctx, projectID, checkpoint.New(wfA, "for A")); err != nil {
		t.Fatal(err)
	}
	if err := db.Checkpoints().Create(ctx, projectID, checkpoint.New(wfB, "for B")); err != nil {
		t.Fatal(err)
	}

	checkpoints, err := db.Checkpoints().List(ctx, projectID, storage.CheckpointFilter{WorkflowID: wfA})
	if err != nil {
		t.Fatal(err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("expected 1 checkpoint for wfA, got %d", len(checkpoints))
	}
	if checkpoints[0].Note != "for A" {
		t.Errorf("Note = %q, want %q", checkpoints[0].Note, "for A")
	}
}

func TestCheckpointStore_ScopedToProject(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	projectA := shared.NewID()
	projectB := shared.NewID()

	if err := db.Checkpoints().Create(ctx, projectA, checkpoint.New(shared.EmptyID, "for A")); err != nil {
		t.Fatal(err)
	}

	checkpoints, err := db.Checkpoints().List(ctx, projectB, storage.CheckpointFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(checkpoints) != 0 {
		t.Errorf("expected 0 checkpoints for project B, got %d", len(checkpoints))
	}
}

func TestEventStore_OrderedByTimestamp(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	// Append in explicit timestamp order (RFC3339 is second-precision;
	// add small pauses or pin timestamps directly on the event struct).
	e1 := event.New(event.TypeProjectInitialized, `{"seq":1}`)
	e1.Timestamp = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	e2 := event.New(event.TypeProjectInitialized, `{"seq":2}`)
	e2.Timestamp = time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
	e3 := event.New(event.TypeProjectInitialized, `{"seq":3}`)
	e3.Timestamp = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, e := range []*event.Event{e3, e1, e2} { // append out of order
		if err := db.Events().Append(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	events, err := db.Events().List(ctx, storage.EventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Payload != `{"seq":1}` {
		t.Errorf("first event should be seq=1, got payload %q", events[0].Payload)
	}
	if events[2].Payload != `{"seq":3}` {
		t.Errorf("last event should be seq=3, got payload %q", events[2].Payload)
	}
}
