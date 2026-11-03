package tui

import (
	"hpdipto/flowterm/activity"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/internal/testsupport"
	"testing"
	"time"
)

func newTrackerSession(t *testing.T) (*db.DB, int64) {
	t.Helper()

	database, _ := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "focus", FocusSessionDuration: 25}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	return database, s.ID
}

func countActivities(t *testing.T, database *db.DB, sessionID int64) int {
	t.Helper()

	rows, err := database.ListSessionActivities(sessionID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}

	return len(rows)
}

func TestTrackerInsertsSpansAsTheyClose(t *testing.T) {
	database, sessionID := newTrackerSession(t)

	tr := newTracker(database, sessionID)
	base := time.Now()

	safari := activity.Sample{App: "com.apple.Safari", Name: "Safari"}
	obsidian := activity.Sample{App: "md.obsidian", Name: "Obsidian"}

	tr.sample(base, safari, db.PhaseFocus)
	tr.sample(base.Add(time.Second), safari, db.PhaseFocus)

	if got := countActivities(t, database, sessionID); got != 0 {
		t.Fatalf("open span persisted early: %d rows", got)
	}

	tr.sample(base.Add(2*time.Second), obsidian, db.PhaseFocus)

	if got := countActivities(t, database, sessionID); got != 1 {
		t.Fatalf("closed span not persisted: %d rows", got)
	}

	tr.sample(base.Add(3*time.Second), obsidian, db.PhaseFocus)
	tr.sample(base.Add(4*time.Second), obsidian, db.PhaseBreak)
	tr.sample(base.Add(6*time.Second), obsidian, db.PhaseBreak)

	if err := tr.stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}

	rows, err := database.ListSessionActivities(sessionID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d spans, want 3", len(rows))
	}

	want := []struct {
		app   string
		phase db.Phase
		dur   time.Duration
	}{
		{"com.apple.Safari", db.PhaseFocus, time.Second},
		{"md.obsidian", db.PhaseFocus, time.Second},
		{"md.obsidian", db.PhaseBreak, 2 * time.Second},
	}

	for i, w := range want {
		if rows[i].App != w.app || rows[i].Phase != w.phase {
			t.Errorf("span %d = %q %q, want %q %q", i, rows[i].App, rows[i].Phase, w.app, w.phase)
		}
		if got := rows[i].EndedAt.Sub(rows[i].StartedAt); got != w.dur {
			t.Errorf("span %d duration = %v, want %v", i, got, w.dur)
		}
	}
}

func TestTrackerIdleEndsSpanAtLastInput(t *testing.T) {
	database, sessionID := newTrackerSession(t)

	tr := newTracker(database, sessionID)
	base := time.Now()
	safari := activity.Sample{App: "com.apple.Safari", Name: "Safari"}

	tr.sample(base, safari, db.PhaseFocus)
	tr.sample(base.Add(10*time.Second), safari, db.PhaseFocus)
	tr.sample(base.Add(70*time.Second), activity.Sample{
		App:  "com.apple.Safari",
		Name: "Safari",
		Idle: 65 * time.Second,
	}, db.PhaseFocus)

	rows, err := database.ListSessionActivities(sessionID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d spans, want 1", len(rows))
	}
	if got := rows[0].EndedAt.Sub(rows[0].StartedAt); got != 5*time.Second {
		t.Errorf("idle span duration = %v, want 5s", got)
	}
	if tr.cur != nil {
		t.Error("idle sample left a span open")
	}
}

func TestTrackerLockEndsSpanAtSampleTime(t *testing.T) {
	database, sessionID := newTrackerSession(t)

	tr := newTracker(database, sessionID)
	base := time.Now()
	safari := activity.Sample{App: "com.apple.Safari", Name: "Safari"}

	tr.sample(base, safari, db.PhaseFocus)
	tr.sample(base.Add(2*time.Second), safari, db.PhaseFocus)
	tr.sample(base.Add(4*time.Second), activity.Sample{
		App:    "com.apple.Safari",
		Name:   "Safari",
		Locked: true,
	}, db.PhaseFocus)

	rows, err := database.ListSessionActivities(sessionID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d spans, want 1", len(rows))
	}
	if got := rows[0].EndedAt.Sub(rows[0].StartedAt); got != 4*time.Second {
		t.Errorf("locked span duration = %v, want 4s", got)
	}
}

func TestTrackerKeepsSpansWhenFlushFails(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "focus", FocusSessionDuration: 25}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	tr := newTracker(database, s.ID)
	base := time.Now()
	safari := activity.Sample{App: "com.apple.Safari", Name: "Safari"}
	obsidian := activity.Sample{App: "md.obsidian", Name: "Obsidian"}

	tr.sample(base, safari, db.PhaseFocus)
	tr.sample(base.Add(time.Second), safari, db.PhaseFocus)
	tr.sample(base.Add(2*time.Second), obsidian, db.PhaseFocus)
	tr.sample(base.Add(3*time.Second), obsidian, db.PhaseFocus)

	if _, err := conn.Exec("DROP TABLE session_activities"); err != nil {
		t.Fatalf("drop table: %v", err)
	}

	tr.sample(base.Add(4*time.Second), safari, db.PhaseFocus)

	if err := tr.stop(); err == nil {
		t.Fatal("expected error while the table is missing")
	}

	if _, err := conn.Exec(`
		CREATE TABLE session_activities (
			id INTEGER PRIMARY KEY,
			session_id INTEGER NOT NULL,
			app TEXT NOT NULL,
			app_name TEXT NOT NULL DEFAULT '',
			phase TEXT NOT NULL DEFAULT '',
			started_at DATETIME NOT NULL,
			ended_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`); err != nil {
		t.Fatalf("recreate table: %v", err)
	}

	if err := tr.stop(); err != nil {
		t.Fatalf("retry flush: %v", err)
	}

	rows, err := database.ListSessionActivities(s.ID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("activities = %d, want 1", len(rows))
	}
	if rows[0].App != "md.obsidian" {
		t.Errorf("activity app = %q, want md.obsidian", rows[0].App)
	}
}
