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
