package tui

import "testing"

func TestSubmitInputStartsTimerAtTags(t *testing.T) {
	m, _ := testModel(t)
	m.currentInput = tags
	setInputValues(&m, map[int]string{
		focusDuration: "25",
		tags:          "go",
	})

	updated, _ := m.submitInput()
	got := updated.(model)

	if got.stage != timerStage {
		t.Errorf("stage = %v, want timerStage",
			got.stage)
	}
}
