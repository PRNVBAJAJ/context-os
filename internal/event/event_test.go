package event_test

import (
	"testing"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestNew_FieldsPopulated(t *testing.T) {
	before := time.Now().UTC()
	e := event.New(event.TypeProjectInitialized, `{"project_id":"abc"}`)
	after := time.Now().UTC()

	if e.ID.IsEmpty() {
		t.Error("ID must not be empty")
	}
	if e.Type != event.TypeProjectInitialized {
		t.Errorf("Type = %q, want %q", e.Type, event.TypeProjectInitialized)
	}
	if e.Payload != `{"project_id":"abc"}` {
		t.Errorf("Payload = %q, want %q", e.Payload, `{"project_id":"abc"}`)
	}
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in expected range [%v, %v]", e.Timestamp, before, after)
	}
	if !e.WorkflowID.IsEmpty() {
		t.Error("WorkflowID should be empty for project-level events")
	}
}

func TestNew_UniqueIDs(t *testing.T) {
	seen := make(map[shared.ID]struct{})
	for range 500 {
		e := event.New(event.TypeProjectInitialized, "{}")
		if _, dup := seen[e.ID]; dup {
			t.Fatalf("duplicate event ID generated: %q", e.ID)
		}
		seen[e.ID] = struct{}{}
	}
}
