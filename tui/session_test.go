package tui

import (
	"github.com/hpdipto/flowterm/activity"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestSessionView(t *testing.T) {
	m := model{
		sessions: []session{
			{name: "Focus", minute: 24, second: 59},
		},
		width:  80,
		height: 24,
	}

	view := m.sessionView()

	if !strings.Contains(view.Content, "Focus") {
		t.Error("expected session name in view")
	}

	for _, row := range strings.Split(bigText("24:59"), "\n") {
		if !strings.Contains(view.Content, row) {
			t.Errorf("expected clock row %q in view", row)
		}
	}

	if !view.AltScreen {
		t.Error("expected alternate screen to be enabled")
	}
}

func TestSessionViewUpdateDoesNothingOutsideTimerStage(t *testing.T) {
	m := model{stage: configStage}

	updatedModel, cmd := m.sessionViewUpdate()
	updated := updatedModel.(model)

	if cmd != nil {
		t.Error("expected no command outside timer stage")
	}

	if updated.stage != configStage {
		t.Error("expected stage to remain unchanged")
	}
}

func TestSessionViewUpdateDecrementsSecond(t *testing.T) {
	m := model{
		stage: timerStage,
		sessions: []session{
			{name: "Focus", minute: 5, second: 10},
		},
	}

	updatedModel, cmd := m.sessionViewUpdate()
	updated := updatedModel.(model)

	if updated.sessions[0].second != 9 {
		t.Errorf("expected 9 seconds, got %d",
			updated.sessions[0].second)
	}

	if cmd == nil {
		t.Error("expected next tick command")
	}
}

func TestSessionViewUpdateRollsOverMinute(t *testing.T) {
	m := model{
		stage: timerStage,
		sessions: []session{
			{name: "Focus", minute: 5, second: 0},
		},
	}

	updatedModel, _ := m.sessionViewUpdate()
	updated := updatedModel.(model)

	if updated.sessions[0].minute != 4 {
		t.Errorf("expected 4 minutes, got %d",
			updated.sessions[0].minute)
	}

	if updated.sessions[0].second != 59 {
		t.Errorf("expected 59 seconds, got %d",
			updated.sessions[0].second)
	}
}

func TestSessionViewUpdateAdvancesToNextSession(t *testing.T) {
	m := model{
		stage: timerStage,
		sessions: []session{
			{name: "Focus", minute: 0, second: 0},
			{name: "Break", minute: 5, second: 0},
		},
		sessionIndex: 0,
	}

	updatedModel, cmd := m.sessionViewUpdate()
	updated := updatedModel.(model)

	if updated.sessionIndex != 1 {
		t.Errorf("expected session index 1, got %d",
			updated.sessionIndex)
	}

	if cmd == nil {
		t.Error("expected next tick command")
	}
}

func TestSessionViewUpdateQuitsWhenSessionsAreExhausted(t *testing.T) {
	m := model{
		stage:        timerStage,
		sessionIndex: 0,
	}

	_, cmd := m.sessionViewUpdate()

	if cmd == nil {
		t.Fatal("expected quit command")
	}

	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected tea.Quit command")
	}
}

func TestSessionViewUpdateWithKeyIgnoresOtherKeys(t *testing.T) {
	m := model{stage: timerStage}

	updatedModel, cmd := m.sessionViewUpdateWithKey(
		tea.KeyPressMsg{Code: 'x'},
	)
	updated := updatedModel.(model)

	if cmd != nil {
		t.Error("expected no command for an unhandled key")
	}

	if updated.stage != timerStage {
		t.Error("expected stage to remain unchanged")
	}
}

func TestSessionViewUpdateWithKeyQuitsWithoutSessions(t *testing.T) {
	m := model{stage: timerStage}

	_, cmd := m.sessionViewUpdateWithKey(
		tea.KeyPressMsg{Code: 'q'},
	)

	if cmd == nil {
		t.Fatal("expected quit command when no sessions exist")
	}

	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("expected tea.Quit command")
	}
}

func TestSessionViewUpdateWithKeyQuitPersistsActivities(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "Focus", FocusSessionDuration: 25}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	m := model{
		stage:    timerStage,
		database: database,
		sessions: []session{{id: s.ID, name: "Focus", minute: 25}},
	}
	m.tracker = newTracker(database, s.ID)

	base := time.Now()
	safari := activity.Sample{App: "com.apple.Safari", Name: "Safari"}
	m.tracker.sample(base, safari, db.PhaseFocus)
	m.tracker.sample(base.Add(2*time.Second), safari, db.PhaseFocus)

	_, cmd := m.sessionViewUpdateWithKey(tea.KeyPressMsg{Code: 'q'})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("cmd() = %T, want tea.QuitMsg", cmd())
	}

	var interrupted bool
	if err := conn.QueryRow("SELECT interrupted FROM sessions WHERE id = ?", s.ID).Scan(&interrupted); err != nil {
		t.Fatalf("read interrupted: %v", err)
	}
	if !interrupted {
		t.Error("quitting mid-session should mark the session interrupted")
	}

	acts, err := database.ListSessionActivities(s.ID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(acts) != 1 {
		t.Fatalf("activities = %d, want 1", len(acts))
	}
	if acts[0].App != "com.apple.Safari" || acts[0].AppName != "Safari" {
		t.Errorf("activity = %q/%q, want com.apple.Safari/Safari", acts[0].App, acts[0].AppName)
	}
	if !acts[0].EndedAt.After(acts[0].StartedAt) {
		t.Errorf("ended_at %v not after started_at %v", acts[0].EndedAt, acts[0].StartedAt)
	}
}

func TestSessionViewUpdateSavesActivitiesOnNaturalCompletion(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "Focus", FocusSessionDuration: 25}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	m := model{
		stage:    timerStage,
		database: database,
		sessions: []session{{id: s.ID, name: "Focus"}},
	}
	m.tracker = newTracker(database, s.ID)

	base := time.Now()
	safari := activity.Sample{App: "com.apple.Safari", Name: "Safari"}
	m.tracker.sample(base, safari, db.PhaseFocus)
	m.tracker.sample(base.Add(2*time.Second), safari, db.PhaseFocus)

	_, cmd := m.sessionViewUpdate()
	if cmd == nil {
		t.Fatal("expected quit command")
	}

	var interrupted bool
	if err := conn.QueryRow("SELECT interrupted FROM sessions WHERE id = ?", s.ID).Scan(&interrupted); err != nil {
		t.Fatalf("read interrupted: %v", err)
	}
	if interrupted {
		t.Error("completing the timer should not mark the session interrupted")
	}

	acts, err := database.ListSessionActivities(s.ID)
	if err != nil {
		t.Fatalf("list activities: %v", err)
	}
	if len(acts) != 1 {
		t.Fatalf("activities = %d, want 1", len(acts))
	}
	if acts[0].App != "com.apple.Safari" {
		t.Errorf("activity app = %q, want com.apple.Safari", acts[0].App)
	}
}

func TestSessionViewUpdateStartsBreakAfterFocus(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "Focus",
		BreakSessionName:     "Break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
	}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	m := model{
		stage:    timerStage,
		database: database,
		sessions: []session{
			{id: s.ID, name: "Focus"},
			{id: s.ID, name: "Break", minute: 5},
		},
	}

	next, cmd := m.sessionViewUpdate()
	if cmd == nil {
		t.Fatal("expected break command")
	}
	if _, ok := cmd().(tea.BatchMsg); !ok {
		t.Fatalf("cmd() = %T, want tea.BatchMsg", cmd())
	}

	got, ok := next.(model)
	if !ok {
		t.Fatalf("model = %T, want tui.model", next)
	}
	if got.sessionIndex != 1 {
		t.Errorf("sessionIndex = %d, want 1", got.sessionIndex)
	}
}

func TestSampleTickStopsWhenBreakSkipped(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "Focus", FocusSessionDuration: 25}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	m := model{
		stage:    timerStage,
		database: database,
		sessions: []session{{id: s.ID, name: "Focus"}},
		tracker:  newTracker(database, s.ID),
	}

	next, cmd := m.sessionViewUpdate()
	if cmd == nil {
		t.Fatal("expected quit command")
	}

	got, ok := next.(model)
	if !ok {
		t.Fatalf("model = %T, want tui.model", next)
	}
	if got.tracker != nil {
		t.Fatal("tracker should be cleared after session end")
	}

	updated, cmd := got.Update(sampleTickMsg(time.Now()))
	if cmd != nil {
		t.Error("expected sampling to stop when no session is left")
	}

	afterTick, ok := updated.(model)
	if !ok {
		t.Fatalf("model = %T, want tui.model", updated)
	}
	if afterTick.tracker != nil {
		t.Error("tracker should stay nil after session end")
	}

	var breakRows int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM session_activities WHERE session_id = ? AND phase = ?",
		s.ID, db.PhaseBreak,
	).Scan(&breakRows); err != nil {
		t.Fatalf("count break activities: %v", err)
	}
	if breakRows != 0 {
		t.Errorf("break activities = %d, want 0", breakRows)
	}
}
