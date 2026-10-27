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
	since := now.Add(-90 * time.Minute)

	completed, err := database.ListSessionStats(since, false)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}
	if len(completed) != 0 {
		t.Fatalf("len(completed) = %d, want 0", len(completed))
	}

	all, err := database.ListSessionStats(since, true)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}

	if len(all) != 1 {
		t.Fatalf("len(all) = %d, want 1", len(all))
	}
	if all[0].FocusSessionName != "beta" {
		t.Errorf("name = %q, want %q", all[0].FocusSessionName, "beta")
	}
	if all[0].Tags != "" {
		t.Errorf("tags = %q, want empty", all[0].Tags)
	}
}

func TestListSessionStatsOrderAndFields(t *testing.T) {
	database, now := seedStats(t)

	stats, err := database.ListSessionStats(now.Add(-3*time.Hour), true)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("len(stats) = %d, want 2", len(stats))
	}

	if stats[0].FocusSessionName != "beta" || stats[1].FocusSessionName != "alpha" {
		t.Fatalf("order = [%q %q], want [beta alpha]", stats[0].FocusSessionName, stats[1].FocusSessionName)
	}
	if !stats[0].Interrupted {
		t.Error("beta interrupted = false, want true")
	}
	if stats[1].Interrupted {
		t.Error("alpha interrupted = true, want false")
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

func TestListSessionStatsAllFilter(t *testing.T) {
	database, now := seedStats(t)
	since := now.Add(-3 * time.Hour)

	completed, err := database.ListSessionStats(since, false)
	if err != nil {
		t.Fatalf("list stats: %v", err)
	}
	if len(completed) != 1 {
		t.Fatalf("len(completed) = %d, want 1", len(completed))
	}
	if completed[0].FocusSessionName != "alpha" {
		t.Errorf("name = %q, want %q", completed[0].FocusSessionName, "alpha")
	}

	all, err := database.ListSessionStats(since, true)
	if err != nil {
		t.Fatalf("list all stats: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}
}
