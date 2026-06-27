package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

// StartWorkflowOptions carries parameters for the StartWorkflow use case.
type StartWorkflowOptions struct {
	RootPath    string
	Name        string
	Description string
}

// StartWorkflow creates a new workflow in running status and persists it.
// It emits a workflow.started event so the audit log reflects the transition.
func StartWorkflow(ctx context.Context, opts StartWorkflowOptions) (*workflow.Workflow, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	w, err := workflow.New(opts.Name, opts.Description)
	if err != nil {
		return nil, err
	}
	if err := w.Start(); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	if err := store.Workflows().Create(ctx, p.ID, w); err != nil {
		return nil, err
	}

	payload := fmt.Sprintf(`{"workflow_id":%q,"workflow_name":%q}`, w.ID, w.Name)
	e := event.New(event.TypeWorkflowStarted, payload)
	e.WorkflowID = w.ID
	if err := store.Events().Append(ctx, e); err != nil {
		return nil, err
	}

	return w, nil
}

// ListWorkflowsOptions carries parameters for the ListWorkflows use case.
type ListWorkflowsOptions struct {
	RootPath string
}

// ListWorkflows returns all workflows for the project, most recent first.
func ListWorkflows(ctx context.Context, opts ListWorkflowsOptions) ([]*workflow.Workflow, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	return store.Workflows().List(ctx, p.ID, storage.WorkflowFilter{})
}

// transitionWorkflowOptions is shared by all status-transition use cases.
type transitionWorkflowOptions struct {
	RootPath  string
	IDPrefix  string
	eventType event.Type
	transition func(*workflow.Workflow) error
}

// transitionWorkflow is the common implementation for Complete, Fail, Pause, Resume.
// It resolves the workflow by ID prefix, calls the domain transition, persists the
// new state, and appends the appropriate event.
func transitionWorkflow(ctx context.Context, opts transitionWorkflowOptions) (*workflow.Workflow, error) {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	w, err := resolveWorkflowByPrefix(ctx, store, p.ID, opts.IDPrefix)
	if err != nil {
		return nil, err
	}

	if err := opts.transition(w); err != nil {
		return nil, err
	}

	if err := store.Workflows().Save(ctx, w); err != nil {
		return nil, err
	}

	payload := fmt.Sprintf(`{"workflow_id":%q}`, w.ID)
	e := event.New(opts.eventType, payload)
	e.WorkflowID = w.ID
	if err := store.Events().Append(ctx, e); err != nil {
		return nil, err
	}

	return w, nil
}

// resolveWorkflowByPrefix finds the unique workflow whose ID starts with prefix.
// Returns CodeNotFound if none match, CodeInvalidInput if multiple match.
func resolveWorkflowByPrefix(ctx context.Context, store storage.Storage, projectID shared.ID, prefix string) (*workflow.Workflow, error) {
	if prefix == "" {
		return nil, shared.NewError(shared.CodeInvalidInput, "workflow ID prefix must not be empty")
	}

	all, err := store.Workflows().List(ctx, projectID, storage.WorkflowFilter{})
	if err != nil {
		return nil, err
	}

	var matches []*workflow.Workflow
	for _, w := range all {
		if strings.HasPrefix(w.ID.String(), prefix) {
			matches = append(matches, w)
		}
	}

	switch len(matches) {
	case 0:
		return nil, shared.NewError(shared.CodeNotFound, "no workflow found with ID prefix "+prefix)
	case 1:
		return matches[0], nil
	default:
		return nil, shared.NewError(shared.CodeInvalidInput,
			fmt.Sprintf("ambiguous prefix %q matches %d workflows — use more characters", prefix, len(matches)))
	}
}

// CompleteWorkflowOptions carries parameters for CompleteWorkflow.
type CompleteWorkflowOptions struct {
	RootPath string
	IDPrefix string
}

// CompleteWorkflow transitions a running workflow to completed.
func CompleteWorkflow(ctx context.Context, opts CompleteWorkflowOptions) (*workflow.Workflow, error) {
	return transitionWorkflow(ctx, transitionWorkflowOptions{
		RootPath:   opts.RootPath,
		IDPrefix:   opts.IDPrefix,
		eventType:  event.TypeWorkflowCompleted,
		transition: (*workflow.Workflow).Complete,
	})
}

// FailWorkflowOptions carries parameters for FailWorkflow.
type FailWorkflowOptions struct {
	RootPath string
	IDPrefix string
}

// FailWorkflow transitions a running workflow to failed.
func FailWorkflow(ctx context.Context, opts FailWorkflowOptions) (*workflow.Workflow, error) {
	return transitionWorkflow(ctx, transitionWorkflowOptions{
		RootPath:   opts.RootPath,
		IDPrefix:   opts.IDPrefix,
		eventType:  event.TypeWorkflowFailed,
		transition: (*workflow.Workflow).Fail,
	})
}

// PauseWorkflowOptions carries parameters for PauseWorkflow.
type PauseWorkflowOptions struct {
	RootPath string
	IDPrefix string
}

// PauseWorkflow transitions a running workflow to paused.
func PauseWorkflow(ctx context.Context, opts PauseWorkflowOptions) (*workflow.Workflow, error) {
	return transitionWorkflow(ctx, transitionWorkflowOptions{
		RootPath:   opts.RootPath,
		IDPrefix:   opts.IDPrefix,
		eventType:  event.TypeWorkflowPaused,
		transition: (*workflow.Workflow).Pause,
	})
}

// ResumeWorkflowOptions carries parameters for ResumeWorkflow.
type ResumeWorkflowOptions struct {
	RootPath string
	IDPrefix string
}

// ResumeWorkflow transitions a paused workflow back to running.
func ResumeWorkflow(ctx context.Context, opts ResumeWorkflowOptions) (*workflow.Workflow, error) {
	return transitionWorkflow(ctx, transitionWorkflowOptions{
		RootPath:   opts.RootPath,
		IDPrefix:   opts.IDPrefix,
		eventType:  event.TypeWorkflowResumed,
		transition: (*workflow.Workflow).Resume,
	})
}

// DeleteWorkflowOptions carries parameters for the DeleteWorkflow use case.
type DeleteWorkflowOptions struct {
	RootPath string
	IDPrefix string
}

// DeleteWorkflow removes a completed or failed workflow from the database.
// It returns CodeInvalidInput if the workflow is still running or paused.
func DeleteWorkflow(ctx context.Context, opts DeleteWorkflowOptions) error {
	p, err := project.Load(opts.RootPath)
	if err != nil {
		return err
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	w, err := resolveWorkflowByPrefix(ctx, store, p.ID, opts.IDPrefix)
	if err != nil {
		return err
	}

	if w.Status == workflow.StatusRunning || w.Status == workflow.StatusPaused {
		return shared.NewError(shared.CodeInvalidInput,
			fmt.Sprintf("cannot delete a %s workflow — use 'context workflow complete' or 'context workflow fail' first", w.Status))
	}

	return store.Workflows().Delete(ctx, w.ID)
}
