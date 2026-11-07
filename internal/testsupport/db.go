package testsupport

import (
	"database/sql"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
)

func OpenDB(t *testing.T) (*db.DB, *sql.DB) {
	t.Helper()

	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() {
		utils.LogError(conn.Close())
	})

	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return db.New(conn), conn
}
