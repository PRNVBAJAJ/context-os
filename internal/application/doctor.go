package application

import (
	"context"
	"path/filepath"

	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
)

// DoctorResult holds the outcome of a runtime health check.
type DoctorResult struct {
	// ProjectName is the name of the project found at rootPath.
	ProjectName string
	// RuntimeVersion is the version recorded in project.yaml.
	RuntimeVersion string
	// DatabaseOK is true when runtime.db was successfully opened and queried.
	DatabaseOK bool
	// EventCount is the total number of events recorded so far.
	EventCount int
	// RecentEvents contains the most recent events, newest last.
	RecentEvents []*event.Event
}

// maxRecentEvents is the number of recent events surfaced by RunDoctor.
const maxRecentEvents = 10

// RunDoctor inspects the runtime health of the project at rootPath and returns
// a DoctorResult. It never returns a partial result on error — callers receive
// either a complete result or an error explaining the first failure.
func RunDoctor(ctx context.Context, rootPath string) (*DoctorResult, error) {
	p, err := project.Load(rootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(rootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	allEvents, err := store.Events().List(ctx, storage.EventFilter{})
	if err != nil {
		return nil, err
	}

	recent, err := store.Events().List(ctx, storage.EventFilter{Limit: maxRecentEvents})
	if err != nil {
		return nil, err
	}

	return &DoctorResult{
		ProjectName:    p.Name,
		RuntimeVersion: p.RuntimeVersion,
		DatabaseOK:     true,
		EventCount:     len(allEvents),
		RecentEvents:   recent,
	}, nil
}
