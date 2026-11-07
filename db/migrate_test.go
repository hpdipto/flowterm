package db_test

import (
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"github.com/hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
)

func TestMigrateIsIdempotent(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	for run := 1; run <= 2; run++ {
		if err := db.Migrate(conn); err != nil {
			t.Fatalf("migrate run %d: %v", run, err)
		}
	}

	// A second Up must not replay the seed migration.
	var count int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM tags WHERE name = ?",
		"work",
	).Scan(&count); err != nil {
		t.Fatalf("count seeded tag: %v", err)
	}
	if count != 1 {
		t.Errorf("seeded tag count = %d, want 1", count)
	}
}

func TestMigrateRemovesOrphanedSessionChildren(t *testing.T) {
	_, conn := testsupport.OpenDB(t)

	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatalf("disable foreign keys: %v", err)
	}

	if _, err := conn.Exec(`
		INSERT INTO session_activities (
			session_id, app, started_at, ended_at, created_at, updated_at
		) VALUES (999, 'app', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert orphan activity: %v", err)
	}

	if _, err := conn.Exec(`
		INSERT INTO session_tags (session_id, tag_id, created_at, updated_at)
		VALUES (999, 1, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`); err != nil {
		t.Fatalf("insert orphan tag link: %v", err)
	}

	if _, err := conn.Exec("DELETE FROM goose_db_version WHERE version_id = 20261106000000"); err != nil {
		t.Fatalf("unmark cascade migration: %v", err)
	}

	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := db.Migrate(conn); err != nil {
		t.Fatalf("re-migrate: %v", err)
	}

	var activities, links int

	if err := conn.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM session_activities WHERE session_id = 999),
			(SELECT COUNT(*) FROM session_tags WHERE session_id = 999)
	`).Scan(&activities, &links); err != nil {
		t.Fatalf("count orphans: %v", err)
	}

	if activities != 0 || links != 0 {
		t.Errorf("orphans = activities %d, links %d; want both 0", activities, links)
	}
}
