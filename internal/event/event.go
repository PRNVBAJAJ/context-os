package event

import (
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

// Type identifies the kind of runtime event.
type Type string

const (
	// TypeProjectInitialized is emitted once when a project is first initialized.
	TypeProjectInitialized Type = "project.initialized"
	// TypeWorkflowStarted is emitted when a workflow transitions to running.
	TypeWorkflowStarted Type = "workflow.started"
	// TypeWorkflowCompleted is emitted when a workflow transitions to completed.
	TypeWorkflowCompleted Type = "workflow.completed"
	// TypeWorkflowFailed is emitted when a workflow transitions to failed.
	TypeWorkflowFailed Type = "workflow.failed"
	// TypeWorkflowPaused is emitted when a workflow transitions to paused.
	TypeWorkflowPaused Type = "workflow.paused"
	// TypeWorkflowResumed is emitted when a workflow transitions back to running from paused.
	TypeWorkflowResumed Type = "workflow.resumed"
)

// Event is an immutable record of something that happened in the runtime.
// Events are append-only and must never be modified after creation.
type Event struct {
	// ID is the unique identifier for this event.
	ID shared.ID
	// WorkflowID is the workflow that triggered the event, if any.
	// It is empty for project-level events.
	WorkflowID shared.ID
	// Type identifies what happened.
	Type Type
	// Payload is a JSON-encoded blob carrying event-specific detail.
	Payload string
	// Timestamp is when the event occurred, in UTC.
	Timestamp time.Time
}

// New constructs an Event with a generated ID and the current UTC timestamp.
// payload must be valid JSON; use "{}" for events with no additional data.
func New(eventType Type, payload string) *Event {
	return &Event{
		ID:        shared.NewID(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}
