package db_test

import (
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"github.com/hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
)

func TestOpenValidPath(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { utils.LogError(conn.Close()) }()

	if conn == nil {
		t.Fatal("expected non-nil connection")
	}

	var one int
	if err := conn.QueryRow("SELECT 1").Scan(&one); err != nil {
		t.Fatalf("query: %v", err)
	}
	if one != 1 {
		t.Errorf("query result = %d, want 1", one)
	}
}

func TestOpenMissingDirectory(t *testing.T) {
	if _, err := db.Open(filepath.Join(t.TempDir(), "missing", "test.db")); err == nil {
		t.Error("expected error for path in missing directory")
	}
}

func TestOpenEnforcesForeignKeys(t *testing.T) {
	_, conn := testsupport.OpenDB(t)

	var enabled int
	if err := conn.QueryRow("PRAGMA foreign_keys").Scan(&enabled); err != nil {
		t.Fatalf("read pragma: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1", enabled)
	}

	_, err := conn.Exec(`
		INSERT INTO session_activities (
			session_id,
			app,
			started_at,
			ended_at,
			created_at,
			updated_at
		) VALUES (999, 'app', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`)
	if err == nil {
		t.Fatal("insert with missing session_id: want foreign key error, got nil")
	}
}
