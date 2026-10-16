package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestUpdateStoresWindowSize(t *testing.T) {
	m := newConfigModel(nil)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	got := updated.(model)

	if got.width != 80 || got.height != 24 {
		t.Errorf("size = %dx%d, want 80x24", got.width, got.height)
	}
}

func TestViewEmptyWhenSessionIndexPastEnd(t *testing.T) {
	m := model{
		stage:        timerStage,
		sessions:     []session{{name: "Focus", minute: 5}},
		sessionIndex: 1,
	}

	if got := m.View(); got.Content != "" {
		t.Errorf("content = %q, want empty", got.Content)
	}
}

func TestTickReturnsTickMsg(t *testing.T) {
	msg := tick()()

	if _, ok := msg.(tickMsg); !ok {
		t.Errorf("msg = %T, want tickMsg", msg)
	}
}

func TestUpdateRoutesKeyPressToConfigStage(t *testing.T) {
	m := newConfigModel(nil)

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("msg = %T, want tea.QuitMsg", msg)
	}
}

func TestUpdateRoutesKeyPressToTimerStage(t *testing.T) {
	m := model{stage: timerStage}

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q'})

	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("msg = %T, want tea.QuitMsg", msg)
	}
}

func TestUpdateIgnoresOtherMsgsAtTimerStage(t *testing.T) {
	m := model{stage: timerStage}

	updated, cmd := m.Update(struct{}{})
	got := updated.(model)

	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
	if got.stage != timerStage {
		t.Errorf("stage = %v, want timerStage",
			got.stage)
	}
}
