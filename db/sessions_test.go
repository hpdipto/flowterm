package db_test

import (
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/internal/testsupport"
	"hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateSession(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
		Interrupted:          true,
	}

	before := time.Now()
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}
	after := time.Now()

	if s.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}
	if s.CreatedAt.Before(before) || s.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", s.CreatedAt, before, after)
	}
	if !s.CreatedAt.Equal(s.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want CreatedAt = %v", s.UpdatedAt, s.CreatedAt)
	}

	got, err := database.GetSession(s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}

	if got.FocusSessionName != s.FocusSessionName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, s.FocusSessionName)
	}
	if got.BreakSessionName != s.BreakSessionName {
		t.Errorf("BreakSessionName = %q, want %q", got.BreakSessionName, s.BreakSessionName)
	}
	if got.FocusSessionDuration != s.FocusSessionDuration {
		t.Errorf("FocusSessionDuration = %d, want %d", got.FocusSessionDuration, s.FocusSessionDuration)
	}
	if got.BreakSessionDuration != s.BreakSessionDuration {
		t.Errorf("BreakSessionDuration = %d, want %d", got.BreakSessionDuration, s.BreakSessionDuration)
	}
	if got.Interrupted != s.Interrupted {
		t.Errorf("Interrupted = %t, want %t", got.Interrupted, s.Interrupted)
	}

	// Timestamps survive the round trip; compare with second precision since
	// DATETIME storage drops sub-second detail.
	if got.CreatedAt.Unix() != s.CreatedAt.Unix() {
		t.Errorf("stored CreatedAt = %v, want %v", got.CreatedAt, s.CreatedAt)
	}
	if got.UpdatedAt.Unix() != s.UpdatedAt.Unix() {
		t.Errorf("stored UpdatedAt = %v, want %v", got.UpdatedAt, s.UpdatedAt)
	}
}

func TestCreateSessionMultipleSessionsGetDistinctIDs(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	first := &db.Session{FocusSessionName: "focus"}
	second := &db.Session{FocusSessionName: "focus"}

	if err := database.CreateSession(first); err != nil {
		t.Fatalf("create first session: %v", err)
	}
	if err := database.CreateSession(second); err != nil {
		t.Fatalf("create second session: %v", err)
	}

	if first.ID == second.ID {
		t.Errorf("both sessions got ID %d, want distinct ids", first.ID)
	}
}

func TestCreateSessionMissingTable(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	s := &db.Session{FocusSessionName: "focus"}

	if err := db.New(conn).CreateSession(s); err == nil {
		t.Error("expected error when sessions table is missing")
	}
	if s.ID != 0 {
		t.Errorf("ID = %d, want 0 after failed insert", s.ID)
	}
}
