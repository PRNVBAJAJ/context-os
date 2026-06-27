package storage

import (
	"context"

	"github.com/PRNVBAJAJ/context-os/internal/event"
)

// EventStore is the append-only persistence interface for runtime events.
// Events must never be modified or deleted after creation.
type EventStore interface {
	// Append persists a new event. The event ID must be unique.
	Append(ctx context.Context, e *event.Event) error
	// List returns events matching filter, ordered by timestamp ascending.
	List(ctx context.Context, filter EventFilter) ([]*event.Event, error)
}
