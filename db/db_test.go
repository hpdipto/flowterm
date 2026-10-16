package db

import (
	"hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
)

func TestOpenValidPath(t *testing.T) {
	conn, err := Open(filepath.Join(t.TempDir(), "test.db"))
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
	if _, err := Open(filepath.Join(t.TempDir(), "missing", "test.db")); err == nil {
		t.Error("expected error for path in missing directory")
	}
}
