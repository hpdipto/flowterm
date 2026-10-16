package tui

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m model) nextInput() (tea.Model, tea.Cmd) {
	m.inputs[m.currentInput].Blur()
	m.currentInput++

	if m.currentInput >= len(m.inputs) {
		return m.startTimer()
	}

	return m.goToInput(m.currentInput)
}

func (m model) goToInput(index int) (tea.Model, tea.Cmd) {
	m.inputs[m.currentInput].Blur()
	m.currentInput = index
	cmd := m.inputs[m.currentInput].Focus()
	return m, cmd
}

func (m model) submitInput() (tea.Model, tea.Cmd) {
	value := strings.TrimSpace(m.inputs[m.currentInput].Value())

	switch m.currentInput {
	case focusName:
		if value == "" {
			m.inputs[focusName].SetValue("Focus")
		}

	case focusDuration:
		minutes, err := strconv.Atoi(value)
		if err != nil || minutes <= 0 {
			m.err = "Enter a positive focus duration."
			return m, nil
		}

	case breakName:
		if value == "" {
			m.inputs[breakName].SetValue("Break")
		}

	case breakDuration:
		minutes, err := strconv.Atoi(value)
		if err != nil || minutes < 0 {
			m.err = "Enter a non-negative break duration."
			return m, nil
		}
	}

	m.err = ""

	if m.currentInput == tags {
		return m.startTimer()
	}

	return m.nextInput()
}
