package application

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

// TrackFileOptions are the inputs for recording a file access.
// Either Filepath or Payload (PostToolUse JSON from stdin) must be set.
type TrackFileOptions struct {
	RootPath string
	// Filepath is the explicit file path to record.
	Filepath string
	// Payload is the raw JSON from a PostToolUse hook (alternative to Filepath).
	// The function extracts the file path from tool_input.file_path.
	Payload string
}

// postToolUsePayload is the shape of the JSON delivered to PostToolUse hooks.
type postToolUsePayload struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

type toolInputWithPath struct {
	FilePath string `json:"file_path"`
	Path     string `json:"path"`
}

// TrackFile records that a file was touched during the active workflow.
// It is a fast, silent operation — errors are intentionally swallowed by callers.
func TrackFile(ctx context.Context, opts TrackFileOptions) error {
	fp := opts.Filepath

	if fp == "" && opts.Payload != "" {
		var p postToolUsePayload
		if err := json.Unmarshal([]byte(opts.Payload), &p); err != nil {
			return nil // malformed payload — silently skip
		}
		// Only track file-touching tools.
		switch p.ToolName {
		case "Read", "Edit", "Write":
		default:
			return nil
		}
		var inp toolInputWithPath
		if err := json.Unmarshal(p.ToolInput, &inp); err != nil {
			return nil
		}
		fp = inp.FilePath
		if fp == "" {
			fp = inp.Path
		}
	}

	if fp == "" {
		return nil
	}

	// Make path relative to project root for portability.
	if filepath.IsAbs(fp) {
		rel, err := filepath.Rel(opts.RootPath, fp)
		if err == nil && !strings.HasPrefix(rel, "..") {
			fp = rel
		}
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	p, err := store.Projects().GetByPath(ctx, opts.RootPath)
	if err != nil {
		return err
	}

	workflows, err := store.Workflows().List(ctx, p.ID, storage.WorkflowFilter{})
	if err != nil {
		return err
	}

	var activeID shared.ID
	for _, w := range workflows {
		if w.Status == workflow.StatusRunning {
			activeID = w.ID
			break
		}
	}
	if activeID.IsEmpty() {
		return nil // no running workflow — nothing to track
	}

	return store.FileAccesses().Record(ctx, activeID, fp)
}

// HotFilesOptions are the inputs for retrieving hot files.
type HotFilesOptions struct {
	RootPath   string
	WorkflowID string // optional — defaults to active running workflow
	Limit      int    // 0 means no cap
}

// HotFile is a file and its access count during a workflow.
type HotFile struct {
	Filepath    string
	AccessCount int
}

// GetHotFiles returns the most-accessed files for the active (or specified) workflow.
func GetHotFiles(ctx context.Context, opts HotFilesOptions) ([]HotFile, error) {
	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	p, err := store.Projects().GetByPath(ctx, opts.RootPath)
	if err != nil {
		return nil, err
	}

	var wfID shared.ID
	if opts.WorkflowID != "" {
		wfID = shared.ID(opts.WorkflowID)
	} else {
		workflows, err := store.Workflows().List(ctx, p.ID, storage.WorkflowFilter{})
		if err != nil {
			return nil, err
		}
		for _, w := range workflows {
			if w.Status == workflow.StatusRunning {
				wfID = w.ID
				break
			}
		}
	}
	if wfID.IsEmpty() {
		return nil, nil
	}

	raw, err := store.FileAccesses().HotFiles(ctx, wfID, opts.Limit)
	if err != nil {
		return nil, err
	}

	out := make([]HotFile, len(raw))
	for i, r := range raw {
		out[i] = HotFile{Filepath: r.Filepath, AccessCount: r.AccessCount}
	}
	return out, nil
}
