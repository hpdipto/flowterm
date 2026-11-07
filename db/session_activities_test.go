package db_test

import (
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"testing"
	"time"
)

func TestCreateSessionActivity(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "focus", FocusSessionDuration: 25}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	start := time.Now().Add(-time.Minute).Truncate(time.Second)

	a := &db.SessionActivity{
		SessionID: s.ID,
		App:       "com.apple.Safari",
		AppName:   "Safari",
		Phase:     db.PhaseFocus,
		StartedAt: start,
		EndedAt:   start.Add(time.Minute),
	}

	before := time.Now()
	if err := database.CreateSessionActivity(a); err != nil {
		t.Fatalf("create activity: %v", err)
	}
	after := time.Now()

	if a.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}
	if a.CreatedAt.Before(before) || a.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", a.CreatedAt, before, after)
	}
	if !a.CreatedAt.Equal(a.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want CreatedAt = %v", a.UpdatedAt, a.CreatedAt)
	}

	activities, err := database.ListSessionActivities(s.ID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("got %d activities, want 1", len(activities))
	}

	got := activities[0]
	if got.App != a.App || got.AppName != a.AppName || got.Phase != a.Phase {
		t.Errorf("activity = %+v, want app %q name %q phase %q", got, a.App, a.AppName, a.Phase)
	}
	if !got.StartedAt.Equal(a.StartedAt) || !got.EndedAt.Equal(a.EndedAt) {
		t.Errorf("span = %v..%v, want %v..%v", got.StartedAt, got.EndedAt, a.StartedAt, a.EndedAt)
	}
}

func TestListSessionActivitiesFiltersAndOrders(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	first := &db.Session{FocusSessionName: "first", FocusSessionDuration: 25}
	second := &db.Session{FocusSessionName: "second", FocusSessionDuration: 25}
	for _, s := range []*db.Session{first, second} {
		if err := database.CreateSession(s); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	base := time.Now().Add(-time.Hour).Truncate(time.Second)

	later := &db.SessionActivity{
		SessionID: first.ID,
		App:       "md.obsidian",
		Phase:     db.PhaseBreak,
		StartedAt: base.Add(10 * time.Minute),
		EndedAt:   base.Add(12 * time.Minute),
	}
	earlier := &db.SessionActivity{
		SessionID: first.ID,
		App:       "com.apple.Safari",
		Phase:     db.PhaseFocus,
		StartedAt: base,
		EndedAt:   base.Add(5 * time.Minute),
	}
	other := &db.SessionActivity{
		SessionID: second.ID,
		App:       "com.apple.Safari",
		Phase:     db.PhaseFocus,
		StartedAt: base,
		EndedAt:   base.Add(time.Minute),
	}

	for _, a := range []*db.SessionActivity{later, earlier, other} {
		if err := database.CreateSessionActivity(a); err != nil {
			t.Fatalf("create activity: %v", err)
		}
	}

	activities, err := database.ListSessionActivities(first.ID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("got %d activities, want 2", len(activities))
	}
	if activities[0].App != earlier.App || activities[1].App != later.App {
		t.Errorf("order = %q, %q, want %q, %q",
			activities[0].App, activities[1].App, earlier.App, later.App)
	}
	if activities[0].Phase != db.PhaseFocus || activities[1].Phase != db.PhaseBreak {
		t.Errorf("phases = %q, %q, want %q, %q",
			activities[0].Phase, activities[1].Phase, db.PhaseFocus, db.PhaseBreak)
	}
}
