package tui

import (
	"database/sql"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"strings"
	"testing"
)

func testModel(t *testing.T) (model, *sql.DB) {
	t.Helper()

	database, conn := testsupport.OpenDB(t)

	return newConfigModel(database), conn
}

func setInputValues(m *model, values map[int]string) {
	for index, value := range values {
		m.inputs[index].SetValue(value)
	}
}

func assertSessionCount(t *testing.T, conn *sql.DB, want int) {
	t.Helper()

	var count int
	if err := conn.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}

	if count != want {
		t.Errorf("session count = %d, want %d", count, want)
	}
}

func TestStartTimerRejectsInvalidFocusDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
	}{
		{name: "empty", duration: ""},
		{name: "non-numeric", duration: "abc"},
		{name: "zero", duration: "0"},
		{name: "negative", duration: "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, conn := testModel(t)
			m.inputs[focusDuration].SetValue(tt.duration)

			updated, _ := m.startTimer()
			got := updated.(model)

			if got.currentInput != focusDuration {
				t.Errorf("currentInput = %d, want %d", got.currentInput, focusDuration)
			}
			if got.stage != configStage {
				t.Errorf("stage = %v, want configStage", got.stage)
			}
			if got.err == "" {
				t.Error("expected validation error")
			}
			assertSessionCount(t, conn, 0)
		})
	}
}

func TestStartTimerDefaultsEmptyFocusName(t *testing.T) {
	m, conn := testModel(t)
	m.inputs[focusName].SetValue("")
	m.inputs[focusDuration].SetValue("25")

	updated, _ := m.startTimer()
	got := updated.(model)

	if len(got.sessions) != 1 {
		t.Fatalf("session count = %d, want 1", len(got.sessions))
	}
	if got.sessions[0].name != "Focus" {
		t.Errorf("focus name = %q, want %q", got.sessions[0].name, "Focus")
	}

	var name string
	err := conn.QueryRow(
		"SELECT focus_session_name FROM sessions",
	).Scan(&name)
	if err != nil {
		t.Fatalf("query focus name: %v", err)
	}
	if name != "Focus" {
		t.Errorf("persisted focus name = %q, want %q", name, "Focus")
	}
}

func TestStartTimerSkipsBreakWithoutName(t *testing.T) {
	m, conn := testModel(t)
	setInputValues(&m, map[int]string{
		focusDuration: "25",
		breakName:     "",
		breakDuration: "5",
	})

	updated, _ := m.startTimer()
	got := updated.(model)

	if len(got.sessions) != 1 {
		t.Fatalf("session count = %d, want 1", len(got.sessions))
	}

	var duration int
	err := conn.QueryRow(
		"SELECT break_session_duration FROM sessions",
	).Scan(&duration)
	if err != nil {
		t.Fatalf("query break duration: %v", err)
	}
	if duration != 0 {
		t.Errorf("persisted break duration = %d, want 0", duration)
	}
}

func TestStartTimerIgnoresInvalidBreakDuration(t *testing.T) {
	m, conn := testModel(t)
	setInputValues(&m, map[int]string{
		focusDuration: "25",
		breakName:     "Rest",
		breakDuration: "abc",
	})

	updated, _ := m.startTimer()
	got := updated.(model)

	if got.err != "" {
		t.Errorf("unexpected error: %q", got.err)
	}
	if len(got.sessions) != 1 {
		t.Fatalf("session count = %d, want 1", len(got.sessions))
	}

	var duration int
	err := conn.QueryRow(
		"SELECT break_session_duration FROM sessions",
	).Scan(&duration)
	if err != nil {
		t.Fatalf("query break duration: %v", err)
	}
	if duration != 0 {
		t.Errorf("persisted break duration = %d, want 0", duration)
	}
}

func TestStartTimerHappyPath(t *testing.T) {
	m, conn := testModel(t)
	setInputValues(&m, map[int]string{
		focusName:     "Deep Work",
		focusDuration: "25",
		breakName:     "Rest",
		breakDuration: "5",
		tags:          "go, tui ,,flow",
	})

	updated, _ := m.startTimer()
	got := updated.(model)

	if got.stage != timerStage {
		t.Errorf("stage = %v, want timerStage", got.stage)
	}
	if got.sessionIndex != 0 {
		t.Errorf("sessionIndex = %d, want 0", got.sessionIndex)
	}
	if len(got.sessions) != 2 {
		t.Fatalf("session count = %d, want 2", len(got.sessions))
	}

	focus := got.sessions[0]
	breakSession := got.sessions[1]

	if focus.name != "Deep Work" {
		t.Errorf("focus name = %q, want %q", focus.name, "Deep Work")
	}
	if focus.minute != 25 {
		t.Errorf("focus duration = %d, want 25", focus.minute)
	}
	if breakSession.name != "Rest" {
		t.Errorf("break name = %q, want %q", breakSession.name, "Rest")
	}
	if breakSession.minute != 5 {
		t.Errorf("break duration = %d, want 5", breakSession.minute)
	}
	if focus.id != breakSession.id {
		t.Errorf("session IDs differ: focus=%d, break=%d", focus.id, breakSession.id)
	}

	var persistedFocusName, persistedBreakName string
	var persistedFocusDuration, persistedBreakDuration int
	var interrupted bool

	err := conn.QueryRow(`
		SELECT focus_session_name, break_session_name,
		       focus_session_duration, break_session_duration, interrupted
		FROM sessions
		WHERE id = ?
	`, focus.id).Scan(
		&persistedFocusName,
		&persistedBreakName,
		&persistedFocusDuration,
		&persistedBreakDuration,
		&interrupted,
	)
	if err != nil {
		t.Fatalf("query persisted session: %v", err)
	}

	if persistedFocusName != "Deep Work" {
		t.Errorf("persisted focus name = %q, want %q", persistedFocusName, "Deep Work")
	}
	if persistedBreakName != "Rest" {
		t.Errorf("persisted break name = %q, want %q", persistedBreakName, "Rest")
	}
	if persistedFocusDuration != 25 {
		t.Errorf("persisted focus duration = %d, want 25", persistedFocusDuration)
	}
	if persistedBreakDuration != 5 {
		t.Errorf("persisted break duration = %d, want 5", persistedBreakDuration)
	}
	if interrupted {
		t.Error("interrupted = true, want false")
	}

	var tagCount int
	err = conn.QueryRow(
		"SELECT COUNT(*) FROM session_tags WHERE session_id = ?",
		focus.id,
	).Scan(&tagCount)
	if err != nil {
		t.Fatalf("count session tags: %v", err)
	}
	if tagCount != 3 {
		t.Errorf("tag count = %d, want 3", tagCount)
	}
}

func TestStartTimerRejectsInvalidTag(t *testing.T) {
	m, conn := testModel(t)
	setInputValues(&m, map[int]string{
		focusDuration: "25",
		tags:          "Bad Tag!",
	})

	updated, _ := m.startTimer()
	got := updated.(model)

	if got.stage != configStage {
		t.Errorf("stage = %v, want configStage", got.stage)
	}
	if !strings.Contains(got.err, "invalid tag name") {
		t.Errorf("error = %q, want invalid tag name error", got.err)
	}

	// Current behavior: the session is created before tag validation fails.
	assertSessionCount(t, conn, 1)
}

func TestStartTimerReportsCreateSessionFailure(t *testing.T) {
	m, conn := testModel(t)

	if err := conn.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	m.inputs[focusDuration].SetValue("25")

	updated, _ := m.startTimer()
	got := updated.(model)

	if got.stage != configStage {
		t.Errorf("stage = %v, want configStage", got.stage)
	}
	if got.err == "" {
		t.Error("expected database error")
	}
}
