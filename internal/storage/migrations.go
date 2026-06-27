package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// migration represents a single schema change.
type migration struct {
	version int
	sql     string
}

// migrations is the ordered list of schema changes for Context OS.
// Each milestone adds new entries here — existing entries are never modified.
var migrations = []migration{
	{
		version: 0,
		sql: `
CREATE TABLE IF NOT EXISTS migration (
    version     INTEGER NOT NULL PRIMARY KEY,
    applied_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS project (
    id              TEXT    NOT NULL PRIMARY KEY,
    name            TEXT    NOT NULL,
    root_path       TEXT    NOT NULL UNIQUE,
    language        TEXT    NOT NULL DEFAULT '',
    runtime_version TEXT    NOT NULL,
    schema_version  INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT    NOT NULL,
    updated_at      TEXT    NOT NULL
);`,
	},
	{
		version: 1,
		sql: `
CREATE TABLE IF NOT EXISTS event (
    id          TEXT    NOT NULL PRIMARY KEY,
    workflow_id TEXT,
    type        TEXT    NOT NULL,
    payload     TEXT    NOT NULL DEFAULT '{}',
    timestamp   TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_event_type ON event(type);`,
	},
	{
		version: 2,
		sql: `
CREATE TABLE IF NOT EXISTS memory (
    id          TEXT NOT NULL PRIMARY KEY,
    project_id  TEXT NOT NULL,
    key         TEXT NOT NULL,
    title       TEXT NOT NULL,
    content     TEXT NOT NULL,
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL,
    UNIQUE(project_id, key)
);

CREATE INDEX IF NOT EXISTS idx_memory_key ON memory(key);`,
	},
	{
		version: 3,
		sql: `
CREATE TABLE IF NOT EXISTS workflow (
    id           TEXT NOT NULL PRIMARY KEY,
    project_id   TEXT NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    started_at   TEXT,
    completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_workflow_project ON workflow(project_id);
CREATE INDEX IF NOT EXISTS idx_workflow_status  ON workflow(status);`,
	},
	{
		version: 4,
		sql: `
CREATE TABLE IF NOT EXISTS checkpoint (
    id          TEXT NOT NULL PRIMARY KEY,
    project_id  TEXT NOT NULL,
    workflow_id TEXT,
    note        TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_checkpoint_project  ON checkpoint(project_id);
CREATE INDEX IF NOT EXISTS idx_checkpoint_workflow ON checkpoint(workflow_id);`,
	},
}

// runMigrations applies any migrations that have not yet been recorded in the
// migration table. Each migration runs inside its own transaction.
func runMigrations(ctx context.Context, db *sql.DB) error {
	// Bootstrap the migration table itself — needed before we can query it.
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS migration (
    version     INTEGER NOT NULL PRIMARY KEY,
    applied_at  TEXT    NOT NULL
)`)
	if err != nil {
		return fmt.Errorf("bootstrap migration table: %w", err)
	}

	for _, m := range migrations {
		applied, err := isMigrationApplied(ctx, db, m.version)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if applied {
			continue
		}

		if err := applyMigration(ctx, db, m); err != nil {
			return fmt.Errorf("apply migration %d: %w", m.version, err)
		}
	}

	return nil
}

func isMigrationApplied(ctx context.Context, db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM migration WHERE version = ?`, version,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func applyMigration(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO migration (version, applied_at) VALUES (?, ?)`,
		m.version,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
