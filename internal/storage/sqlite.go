package storage

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/PRNVBAJAJ/context-os/internal/checkpoint"
	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
	_ "modernc.org/sqlite" // registers the "sqlite" driver with database/sql
)

// Open opens or creates the SQLite database at path, runs all pending schema
// migrations, and returns a ready-to-use Storage. The caller must call Close()
// when the Storage is no longer needed.
func Open(ctx context.Context, path string) (Storage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to open sqlite database", err)
	}

	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, shared.Wrap(shared.CodeInternal, "failed to connect to sqlite database", err)
	}

	if err := runMigrations(ctx, db); err != nil {
		_ = db.Close()
		return nil, shared.Wrap(shared.CodeInternal, "schema migration failed", err)
	}

	return &sqliteStorage{db: db}, nil
}

// sqliteStorage implements Storage using a SQLite database.
type sqliteStorage struct {
	db *sql.DB
}

func (s *sqliteStorage) Projects() ProjectStore {
	return &sqliteProjectStore{db: s.db}
}

func (s *sqliteStorage) Events() EventStore {
	return &sqliteEventStore{db: s.db}
}

func (s *sqliteStorage) Memories() MemoryStore {
	return &sqliteMemoryStore{db: s.db}
}

func (s *sqliteStorage) Workflows() WorkflowStore {
	return &sqliteWorkflowStore{db: s.db}
}

func (s *sqliteStorage) Checkpoints() CheckpointStore {
	return &sqliteCheckpointStore{db: s.db}
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

// sqliteEventStore implements EventStore against the event table.
type sqliteEventStore struct {
	db *sql.DB
}

func (s *sqliteEventStore) Append(ctx context.Context, e *event.Event) error {
	workflowID := ""
	if !e.WorkflowID.IsEmpty() {
		workflowID = e.WorkflowID.String()
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO event (id, workflow_id, type, payload, timestamp)
VALUES (?, NULLIF(?, ''), ?, ?, ?)`,
		e.ID.String(),
		workflowID,
		string(e.Type),
		e.Payload,
		e.Timestamp.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to append event", err)
	}
	return nil
}

func (s *sqliteEventStore) List(ctx context.Context, filter EventFilter) ([]*event.Event, error) {
	q := `SELECT id, COALESCE(workflow_id, ''), type, payload, timestamp FROM event ORDER BY timestamp ASC`
	args := []any{}

	if filter.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to list events", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*event.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "error iterating event rows", err)
	}
	return events, nil
}

func scanEvent(rows *sql.Rows) (*event.Event, error) {
	var (
		id           string
		workflowID   string
		eventType    string
		payload      string
		timestampStr string
	)

	if err := rows.Scan(&id, &workflowID, &eventType, &payload, &timestampStr); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan event row", err)
	}

	ts, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse event timestamp", err)
	}

	e := &event.Event{
		ID:        shared.ID(id),
		Type:      event.Type(eventType),
		Payload:   payload,
		Timestamp: ts.UTC(),
	}
	if workflowID != "" {
		e.WorkflowID = shared.ID(workflowID)
	}
	return e, nil
}

// sqliteProjectStore implements ProjectStore against the project table.
type sqliteProjectStore struct {
	db *sql.DB
}

func (s *sqliteProjectStore) Create(ctx context.Context, p *project.Project) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO project (id, name, root_path, language, runtime_version, schema_version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID.String(),
		p.Name,
		p.RootPath,
		p.Language,
		p.RuntimeVersion,
		p.SchemaVersion,
		p.CreatedAt.UTC().Format(time.RFC3339),
		p.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		// SQLite UNIQUE constraint violation on root_path.
		if isSQLiteUniqueViolation(err) {
			return shared.NewError(shared.CodeConflict, "a project at this path already exists in the database")
		}
		return shared.Wrap(shared.CodeInternal, "failed to create project record", err)
	}
	return nil
}

func (s *sqliteProjectStore) GetByPath(ctx context.Context, rootPath string) (*project.Project, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, name, root_path, language, runtime_version, schema_version, created_at, updated_at
FROM project
WHERE root_path = ?`, rootPath)

	return scanProject(row)
}

func scanProject(row *sql.Row) (*project.Project, error) {
	var (
		id             string
		name           string
		rootPath       string
		language       string
		runtimeVersion string
		schemaVersion  int
		createdAtStr   string
		updatedAtStr   string
	)

	err := row.Scan(&id, &name, &rootPath, &language, &runtimeVersion, &schemaVersion, &createdAtStr, &updatedAtStr)
	if err == sql.ErrNoRows {
		return nil, shared.NewError(shared.CodeNotFound, "project not found")
	}
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan project row", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse project created_at", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse project updated_at", err)
	}

	return &project.Project{
		ID:             shared.ID(id),
		Name:           name,
		RootPath:       rootPath,
		Language:       language,
		RuntimeVersion: runtimeVersion,
		SchemaVersion:  schemaVersion,
		CreatedAt:      createdAt.UTC(),
		UpdatedAt:      updatedAt.UTC(),
	}, nil
}

// isSQLiteUniqueViolation detects SQLite UNIQUE constraint errors.
// modernc.org/sqlite surfaces these as an error whose message contains
// "UNIQUE constraint failed".
func isSQLiteUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// sqliteMemoryStore implements MemoryStore against the memory table.
type sqliteMemoryStore struct {
	db *sql.DB
}

func (s *sqliteMemoryStore) Add(ctx context.Context, projectID shared.ID, m *memory.Memory) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO memory (id, project_id, key, title, content, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID.String(),
		projectID.String(),
		m.Key,
		m.Title,
		m.Content,
		m.CreatedAt.UTC().Format(time.RFC3339),
		m.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		if isSQLiteUniqueViolation(err) {
			return shared.NewError(shared.CodeConflict, "a memory with this key already exists for this project")
		}
		return shared.Wrap(shared.CodeInternal, "failed to add memory", err)
	}
	return nil
}

func (s *sqliteMemoryStore) List(ctx context.Context, projectID shared.ID, filter MemoryFilter) ([]*memory.Memory, error) {
	q := `SELECT id, key, title, content, created_at, updated_at FROM memory WHERE project_id = ? ORDER BY created_at ASC`
	args := []any{projectID.String()}

	if filter.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to list memories", err)
	}
	defer func() { _ = rows.Close() }()

	var memories []*memory.Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "error iterating memory rows", err)
	}
	return memories, nil
}

func (s *sqliteMemoryStore) GetByKey(ctx context.Context, projectID shared.ID, key string) (*memory.Memory, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, key, title, content, created_at, updated_at
FROM memory
WHERE project_id = ? AND key = ?`, projectID.String(), key)

	m, err := scanMemoryRow(row)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func scanMemory(rows *sql.Rows) (*memory.Memory, error) {
	var (
		id           string
		key          string
		title        string
		content      string
		createdAtStr string
		updatedAtStr string
	)
	if err := rows.Scan(&id, &key, &title, &content, &createdAtStr, &updatedAtStr); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan memory row", err)
	}
	return parseMemoryFields(id, key, title, content, createdAtStr, updatedAtStr)
}

func scanMemoryRow(row *sql.Row) (*memory.Memory, error) {
	var (
		id           string
		key          string
		title        string
		content      string
		createdAtStr string
		updatedAtStr string
	)
	err := row.Scan(&id, &key, &title, &content, &createdAtStr, &updatedAtStr)
	if err == sql.ErrNoRows {
		return nil, shared.NewError(shared.CodeNotFound, "memory not found")
	}
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan memory row", err)
	}
	return parseMemoryFields(id, key, title, content, createdAtStr, updatedAtStr)
}

func parseMemoryFields(id, key, title, content, createdAtStr, updatedAtStr string) (*memory.Memory, error) {
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse memory created_at", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse memory updated_at", err)
	}
	return &memory.Memory{
		ID:        shared.ID(id),
		Key:       key,
		Title:     title,
		Content:   content,
		CreatedAt: createdAt.UTC(),
		UpdatedAt: updatedAt.UTC(),
	}, nil
}

// sqliteWorkflowStore implements WorkflowStore against the workflow table.
type sqliteWorkflowStore struct {
	db *sql.DB
}

func (s *sqliteWorkflowStore) Create(ctx context.Context, projectID shared.ID, w *workflow.Workflow) error {
	startedAt := ""
	if w.StartedAt != nil {
		startedAt = w.StartedAt.UTC().Format(time.RFC3339)
	}
	completedAt := ""
	if w.CompletedAt != nil {
		completedAt = w.CompletedAt.UTC().Format(time.RFC3339)
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO workflow (id, project_id, name, description, status, created_at, updated_at, started_at, completed_at)
VALUES (?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''))`,
		w.ID.String(),
		projectID.String(),
		w.Name,
		w.Description,
		string(w.Status),
		w.CreatedAt.UTC().Format(time.RFC3339),
		w.UpdatedAt.UTC().Format(time.RFC3339),
		startedAt,
		completedAt,
	)
	if err != nil {
		if isSQLiteUniqueViolation(err) {
			return shared.NewError(shared.CodeConflict, "a workflow with this ID already exists")
		}
		return shared.Wrap(shared.CodeInternal, "failed to create workflow", err)
	}
	return nil
}

func (s *sqliteWorkflowStore) Save(ctx context.Context, w *workflow.Workflow) error {
	startedAt := ""
	if w.StartedAt != nil {
		startedAt = w.StartedAt.UTC().Format(time.RFC3339)
	}
	completedAt := ""
	if w.CompletedAt != nil {
		completedAt = w.CompletedAt.UTC().Format(time.RFC3339)
	}

	result, err := s.db.ExecContext(ctx, `
UPDATE workflow
SET status = ?, updated_at = ?, started_at = NULLIF(?, ''), completed_at = NULLIF(?, '')
WHERE id = ?`,
		string(w.Status),
		w.UpdatedAt.UTC().Format(time.RFC3339),
		startedAt,
		completedAt,
		w.ID.String(),
	)
	if err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to save workflow", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to check rows affected", err)
	}
	if rows == 0 {
		return shared.NewError(shared.CodeNotFound, "workflow not found")
	}
	return nil
}

func (s *sqliteWorkflowStore) GetByID(ctx context.Context, id shared.ID) (*workflow.Workflow, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, name, description, status, created_at, updated_at, started_at, completed_at
FROM workflow
WHERE id = ?`, id.String())

	return scanWorkflowRow(row)
}

func (s *sqliteWorkflowStore) List(ctx context.Context, projectID shared.ID, filter WorkflowFilter) ([]*workflow.Workflow, error) {
	q := `
SELECT id, name, description, status, created_at, updated_at, started_at, completed_at
FROM workflow
WHERE project_id = ?
ORDER BY created_at DESC`
	args := []any{projectID.String()}

	if filter.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to list workflows", err)
	}
	defer func() { _ = rows.Close() }()

	var workflows []*workflow.Workflow
	for rows.Next() {
		w, err := scanWorkflow(rows)
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, w)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "error iterating workflow rows", err)
	}
	return workflows, nil
}

func scanWorkflowRow(row *sql.Row) (*workflow.Workflow, error) {
	var (
		id           string
		name         string
		description  string
		status       string
		createdAtStr string
		updatedAtStr string
		startedAtStr sql.NullString
		completedStr sql.NullString
	)
	err := row.Scan(&id, &name, &description, &status, &createdAtStr, &updatedAtStr, &startedAtStr, &completedStr)
	if err == sql.ErrNoRows {
		return nil, shared.NewError(shared.CodeNotFound, "workflow not found")
	}
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan workflow row", err)
	}
	return parseWorkflowFields(id, name, description, status, createdAtStr, updatedAtStr, startedAtStr, completedStr)
}

func scanWorkflow(rows *sql.Rows) (*workflow.Workflow, error) {
	var (
		id           string
		name         string
		description  string
		status       string
		createdAtStr string
		updatedAtStr string
		startedAtStr sql.NullString
		completedStr sql.NullString
	)
	if err := rows.Scan(&id, &name, &description, &status, &createdAtStr, &updatedAtStr, &startedAtStr, &completedStr); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan workflow row", err)
	}
	return parseWorkflowFields(id, name, description, status, createdAtStr, updatedAtStr, startedAtStr, completedStr)
}

func parseWorkflowFields(id, name, description, status, createdAtStr, updatedAtStr string, startedAtStr, completedStr sql.NullString) (*workflow.Workflow, error) {
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse workflow created_at", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse workflow updated_at", err)
	}

	w := &workflow.Workflow{
		ID:          shared.ID(id),
		Name:        name,
		Description: description,
		Status:      workflow.Status(status),
		CreatedAt:   createdAt.UTC(),
		UpdatedAt:   updatedAt.UTC(),
	}

	if startedAtStr.Valid && startedAtStr.String != "" {
		t, err := time.Parse(time.RFC3339, startedAtStr.String)
		if err != nil {
			return nil, shared.Wrap(shared.CodeInternal, "failed to parse workflow started_at", err)
		}
		utc := t.UTC()
		w.StartedAt = &utc
	}

	if completedStr.Valid && completedStr.String != "" {
		t, err := time.Parse(time.RFC3339, completedStr.String)
		if err != nil {
			return nil, shared.Wrap(shared.CodeInternal, "failed to parse workflow completed_at", err)
		}
		utc := t.UTC()
		w.CompletedAt = &utc
	}

	return w, nil
}

// sqliteCheckpointStore implements CheckpointStore against the checkpoint table.
type sqliteCheckpointStore struct {
	db *sql.DB
}

func (s *sqliteCheckpointStore) Create(ctx context.Context, projectID shared.ID, cp *checkpoint.Checkpoint) error {
	workflowID := ""
	if !cp.WorkflowID.IsEmpty() {
		workflowID = cp.WorkflowID.String()
	}

	_, err := s.db.ExecContext(ctx, `
INSERT INTO checkpoint (id, project_id, workflow_id, note, created_at)
VALUES (?, ?, NULLIF(?, ''), ?, ?)`,
		cp.ID.String(),
		projectID.String(),
		workflowID,
		cp.Note,
		cp.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return shared.Wrap(shared.CodeInternal, "failed to create checkpoint", err)
	}
	return nil
}

func (s *sqliteCheckpointStore) List(ctx context.Context, projectID shared.ID, filter CheckpointFilter) ([]*checkpoint.Checkpoint, error) {
	q := `
SELECT id, COALESCE(workflow_id, ''), note, created_at
FROM checkpoint
WHERE project_id = ?`
	args := []any{projectID.String()}

	if !filter.WorkflowID.IsEmpty() {
		q += ` AND workflow_id = ?`
		args = append(args, filter.WorkflowID.String())
	}

	q += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		q += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to list checkpoints", err)
	}
	defer func() { _ = rows.Close() }()

	var checkpoints []*checkpoint.Checkpoint
	for rows.Next() {
		cp, err := scanCheckpoint(rows)
		if err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "error iterating checkpoint rows", err)
	}
	return checkpoints, nil
}

func scanCheckpoint(rows *sql.Rows) (*checkpoint.Checkpoint, error) {
	var (
		id           string
		workflowID   string
		note         string
		createdAtStr string
	)
	if err := rows.Scan(&id, &workflowID, &note, &createdAtStr); err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to scan checkpoint row", err)
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, shared.Wrap(shared.CodeInternal, "failed to parse checkpoint created_at", err)
	}
	cp := &checkpoint.Checkpoint{
		ID:        shared.ID(id),
		Note:      note,
		CreatedAt: createdAt.UTC(),
	}
	if workflowID != "" {
		cp.WorkflowID = shared.ID(workflowID)
	}
	return cp, nil
}
