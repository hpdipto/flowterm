package db_test

import (
	"database/sql"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/internal/testsupport"
	"slices"
	"strings"
	"testing"
	"time"
)

const (
	statFocusDuration = 25
	statBreakDuration = 5
)

func insertSessionAt(t *testing.T, conn *sql.DB, name string, createdAt time.Time, interrupted bool) int64 {
	t.Helper()

	res, err := conn.Exec(`
		INSERT INTO sessions (
			focus_session_name,
			break_session_name,
			focus_session_duration,
			break_session_duration,
			interrupted,
			created_at,
			updated_at
		) VALUES (?, 'break', ?, ?, ?, ?, ?)
	`, name, statFocusDuration, statBreakDuration, interrupted, createdAt, createdAt)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	return id
}

func attachTag(t *testing.T, database *db.DB, sessionID int64, name string) {
	t.Helper()

	tag, err := database.GetOrCreateTag(name)
	if err != nil {
		t.Fatalf("get or create tag %q: %v", name, err)
	}

	if err := database.AddTagToSession(sessionID, tag.ID); err != nil {
		t.Fatalf("add tag %q to session: %v", name, err)
	}
}

// seedStats inserts "alpha" (2 hours old, tags work and alpha) and "beta"
// (1 hour old, interrupted, no tags).
func seedStats(t *testing.T) (*db.DB, time.Time) {
	t.Helper()

	database, conn := testsupport.OpenDB(t)
	now := time.Now()

	alpha := insertSessionAt(t, conn, "alpha", now.Add(-2*time.Hour), false)
	attachTag(t, database, alpha, "work")
	attachTag(t, database, alpha, "alpha")

	insertSessionAt(t, conn, "beta", now.Add(-time.Hour), true)

	return database, now
}

func TestListSessionStatsWindow(t *testing.T) {
	database, now := seedStats(t)

	stats, err := database.ListSessionStats(now.Add(-90*time.Minute), false)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf("len(stats) = %d, want 1", len(stats))
	}
	if stats[0].FocusSessionName != "beta" {
		t.Errorf("name = %q, want %q", stats[0].FocusSessionName, "beta")
	}
	if stats[0].Tags != "" {
		t.Errorf("tags = %q, want empty", stats[0].Tags)
	}
}

func TestListSessionStatsOrderAndFields(t *testing.T) {
	database, now := seedStats(t)

	stats, err := database.ListSessionStats(now.Add(-3*time.Hour), false)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("len(stats) = %d, want 2", len(stats))
	}

	if stats[0].FocusSessionName != "beta" || stats[1].FocusSessionName != "alpha" {
		t.Fatalf("order = [%q %q], want [beta alpha]", stats[0].FocusSessionName, stats[1].FocusSessionName)
	}

	alpha := stats[1]
	if alpha.FocusSessionDuration != statFocusDuration {
		t.Errorf("focus duration = %d, want %d", alpha.FocusSessionDuration, statFocusDuration)
	}
	if alpha.BreakSessionName != "break" {
		t.Errorf("break name = %q, want %q", alpha.BreakSessionName, "break")
	}
	if alpha.BreakSessionDuration != statBreakDuration {
		t.Errorf("break duration = %d, want %d", alpha.BreakSessionDuration, statBreakDuration)
	}

	tags := strings.Split(alpha.Tags, ",")
	slices.Sort(tags)
	if !slices.Equal(tags, []string{"alpha", "work"}) {
		t.Errorf("tags = %q, want alpha and work", alpha.Tags)
	}
}

func TestListSessionStatsInterruptedFilter(t *testing.T) {
	database, now := seedStats(t)
	since := now.Add(-3 * time.Hour)

	all, err := database.ListSessionStats(since, false)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}

	onlyInterrupted, err := database.ListSessionStats(since, true)
	if err != nil {
		t.Fatalf("list interrupted stats: %v", err)
	}
	if len(onlyInterrupted) != 1 {
		t.Fatalf("len(onlyInterrupted) = %d, want 1", len(onlyInterrupted))
	}
	if onlyInterrupted[0].FocusSessionName != "beta" {
		t.Errorf("name = %q, want %q", onlyInterrupted[0].FocusSessionName, "beta")
	}
}
