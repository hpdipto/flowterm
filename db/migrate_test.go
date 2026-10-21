package db_test

import (
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/utils"
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
