package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewModel(t *testing.T) {
	m := newConfigModel(nil)

	if m.stage != configStage {
		t.Errorf("expected config stage, got %v", m.stage)
	}

	if len(m.inputs) != 5 {
		t.Fatalf("expected 5 inputs, got %d", len(m.inputs))
	}

	if m.inputs[focusName].Placeholder != "Press Tab to use default name (Focus)" {
		t.Errorf("unexpected focus name placeholder")
	}
}

func TestConfigView(t *testing.T) {
	m := newConfigModel(nil)
	m.width = 100
	m.height = 30

	view := m.configView()

	if !strings.Contains(view.Content, "Configure your session") {
		t.Error("expected configuration heading in view")
	}

	if !view.AltScreen {
		t.Error("expected alternate screen to be enabled")
	}
}

func TestConfigTabSetsDefaultFocusName(t *testing.T) {
	m := newConfigModel(nil)

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyTab},
	)
	updated := updatedModel.(model)

	if updated.inputs[focusName].Value() != "Focus" {
		t.Errorf("expected default focus name, got %q",
			updated.inputs[focusName].Value())
	}
}

func TestConfigEscapeSkipsBreak(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = breakName

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEscape},
	)
	updated := updatedModel.(model)

	if updated.inputs[breakName].Value() != "" {
		t.Error("expected break name to be empty")
	}

	if updated.inputs[breakDuration].Value() != "0" {
		t.Errorf("expected break duration 0, got %q",
			updated.inputs[breakDuration].Value())
	}

	if updated.currentInput != tags {
		t.Errorf("expected tags input, got %v", updated.currentInput)
	}
}

func TestConfigCtrlCQuits(t *testing.T) {
	m := newConfigModel(nil)

	_, cmd := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl},
	)

	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestConfigTabSetsDefaultBreakName(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = breakName

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyTab},
	)
	updated := updatedModel.(model)

	if updated.inputs[breakName].Value() != "Break" {
		t.Errorf("expected default break name, got %q",
			updated.inputs[breakName].Value())
	}

	if updated.currentInput != breakDuration {
		t.Errorf("expected break duration input, got %v",
			updated.currentInput)
	}
}

func TestConfigTabDoesNothingOnDurationInput(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = focusDuration
	m.inputs[focusDuration].SetValue("25")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyTab},
	)
	updated := updatedModel.(model)

	if updated.inputs[focusDuration].Value() != "25" {
		t.Errorf("expected duration to remain 25, got %q",
			updated.inputs[focusDuration].Value())
	}

	if updated.currentInput != focusDuration {
		t.Errorf("expected focus duration input to remain active, got %v",
			updated.currentInput)
	}
}

func TestConfigEnterAdvancesAfterValidFocusDuration(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = focusDuration
	m.inputs[focusDuration].SetValue("25")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEnter},
	)
	updated := updatedModel.(model)

	if updated.currentInput != breakName {
		t.Errorf("expected break name input, got %v",
			updated.currentInput)
	}

	if updated.err != "" {
		t.Errorf("expected no error, got %q", updated.err)
	}
}

func TestConfigEnterRejectsInvalidFocusDuration(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = focusDuration
	m.inputs[focusDuration].SetValue("0")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEnter},
	)
	updated := updatedModel.(model)

	if updated.currentInput != focusDuration {
		t.Errorf("expected to remain on focus duration, got %v",
			updated.currentInput)
	}

	if updated.err != "Enter a positive focus duration." {
		t.Errorf("unexpected error: %q", updated.err)
	}
}

func TestConfigEnterRejectsNegativeBreakDuration(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = breakDuration
	m.inputs[breakDuration].SetValue("-1")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEnter},
	)
	updated := updatedModel.(model)

	if updated.err != "Enter a non-negative break duration." {
		t.Errorf("unexpected error: %q", updated.err)
	}

	if updated.currentInput != breakDuration {
		t.Errorf("expected to remain on break duration, got %v",
			updated.currentInput)
	}
}

func TestConfigEnterDefaultsEmptyFocusName(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = focusName
	m.inputs[focusName].SetValue("")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEnter},
	)
	updated := updatedModel.(model)

	if updated.inputs[focusName].Value() != "Focus" {
		t.Errorf("expected default focus name, got %q",
			updated.inputs[focusName].Value())
	}

	if updated.currentInput != focusDuration {
		t.Errorf("expected focus duration input, got %v",
			updated.currentInput)
	}
}

func TestConfigEnterDefaultsEmptyBreakName(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = breakName
	m.inputs[breakName].SetValue("")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEnter},
	)
	updated := updatedModel.(model)

	if updated.inputs[breakName].Value() != "Break" {
		t.Errorf("expected default break name, got %q",
			updated.inputs[breakName].Value())
	}

	if updated.currentInput != breakDuration {
		t.Errorf("expected break duration input, got %v",
			updated.currentInput)
	}
}

func TestConfigEnterRejectsNonNumericBreakDuration(t *testing.T) {
	m := newConfigModel(nil)
	m.currentInput = breakDuration
	m.inputs[breakDuration].SetValue("abc")

	updatedModel, _ := m.configViewUpdateWithKey(
		tea.KeyPressMsg{Code: tea.KeyEnter},
	)
	updated := updatedModel.(model)

	if updated.err != "Enter a non-negative break duration." {
		t.Errorf("unexpected error: %q", updated.err)
	}

	if updated.currentInput != breakDuration {
		t.Errorf("expected to remain on break duration, got %v",
			updated.currentInput)
	}
}
