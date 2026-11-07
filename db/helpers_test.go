package db_test

import (
	"github.com/hpdipto/flowterm/config"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"github.com/hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
)

func TestCreateSessionFromConfig(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	c := config.Config{
		FocusSessionName: "Deep Work",
		FocusMinute:      25,
		BreakSessionName: "Rest",
		BreakMinute:      5,
		Tags:             []string{"go", "work"},
	}

	session, err := db.CreateSessionFromConfig(database, c)
	if err != nil {
		t.Fatalf("create session from config: %v", err)
	}

	if session.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}
	if session.Interrupted {
		t.Error("Interrupted = true, want false")
	}

	got, err := database.GetSession(session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.FocusSessionName != c.FocusSessionName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, c.FocusSessionName)
	}
	if got.BreakSessionName != c.BreakSessionName {
		t.Errorf("BreakSessionName = %q, want %q", got.BreakSessionName, c.BreakSessionName)
	}
	if got.FocusSessionDuration != c.FocusMinute {
		t.Errorf("FocusSessionDuration = %d, want %d", got.FocusSessionDuration, c.FocusMinute)
	}
	if got.BreakSessionDuration != c.BreakMinute {
		t.Errorf("BreakSessionDuration = %d, want %d", got.BreakSessionDuration, c.BreakMinute)
	}

	var count int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM session_tags WHERE session_id = ?",
		session.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count session tags: %v", err)
	}
	if count != len(c.Tags) {
		t.Errorf("tag link count = %d, want %d", count, len(c.Tags))
	}

	// The seeded "work" tag is reused, not duplicated, and linked to the
	// session.
	workTag, err := database.GetTagByName("work")
	if err != nil {
		t.Fatalf("get seeded tag: %v", err)
	}
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM session_tags WHERE session_id = ? AND tag_id = ?",
		session.ID,
		workTag.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count seeded tag links: %v", err)
	}
	if count != 1 {
		t.Errorf("seeded tag link count = %d, want 1", count)
	}

	if err := conn.QueryRow("SELECT COUNT(*) FROM tags").Scan(&count); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if count != 2 {
		t.Errorf("tag count = %d, want 2 (seed work + go)", count)
	}
}

func TestCreateSessionFromConfigInvalidTag(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	session, err := db.CreateSessionFromConfig(database, config.Config{
		FocusSessionName: "Deep Work",
		FocusMinute:      25,
		Tags:             []string{"go", "Bad Tag!"},
	})
	if err == nil {
		t.Fatal("expected invalid tag error")
	}
	if session != nil {
		t.Errorf("session = %+v, want nil", session)
	}

	// Current behavior: the session row is inserted before tag validation, and
	// the tags processed before the invalid one stay linked.
	var sessionCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 1 {
		t.Errorf("session count = %d, want 1", sessionCount)
	}

	var linkCount int
	if err := conn.QueryRow("SELECT COUNT(*) FROM session_tags").Scan(&linkCount); err != nil {
		t.Fatalf("count session tags: %v", err)
	}
	if linkCount != 1 {
		t.Errorf("tag link count = %d, want 1", linkCount)
	}
}

func TestCreateSessionFromConfigMissingTable(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	session, err := db.CreateSessionFromConfig(db.New(conn), config.Config{
		FocusSessionName: "Deep Work",
	})
	if err == nil {
		t.Fatal("expected error when sessions table is missing")
	}
	if session != nil {
		t.Errorf("session = %+v, want nil", session)
	}
}

func TestCreateSessionFromConfigAddTagFailure(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	// Force every session_tags insert to fail, so the failure happens after the
	// session row and the tag exist.
	if _, err := conn.Exec(`
		CREATE TRIGGER reject_link_insert
		BEFORE INSERT ON session_tags
		BEGIN
			SELECT RAISE(FAIL, 'insert rejected');
		END;
	`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	session, err := db.CreateSessionFromConfig(database, config.Config{
		FocusSessionName: "Deep Work",
		Tags:             []string{"go"},
	})
	if err == nil {
		t.Fatal("expected error when tag link insert fails")
	}
	if session != nil {
		t.Errorf("session = %+v, want nil", session)
	}
}

func TestCloseSession(t *testing.T) {
	tests := []struct {
		name        string
		interrupted bool
	}{
		{name: "interrupted", interrupted: true},
		{name: "completed", interrupted: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, _ := testsupport.OpenDB(t)

			session := &db.Session{FocusSessionName: "focus"}
			if err := database.CreateSession(session); err != nil {
				t.Fatalf("create session: %v", err)
			}

			if err := db.CloseSession(database, session.ID, tt.interrupted); err != nil {
				t.Fatalf("close session: %v", err)
			}

			got, err := database.GetSession(session.ID)
			if err != nil {
				t.Fatalf("get session: %v", err)
			}
			if got.Interrupted != tt.interrupted {
				t.Errorf("Interrupted = %t, want %t", got.Interrupted, tt.interrupted)
			}
			if got.FocusSessionName != session.FocusSessionName {
				t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, session.FocusSessionName)
			}
		})
	}
}
