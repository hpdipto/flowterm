package tui

import (
	"strings"
	"testing"

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

	if !strings.Contains(view.Content, "24:59") {
		t.Error("expected remaining time in view")
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
