package tui

import (
	"fmt"
	"hpdipto/flowterm/db"
	"image/color"
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type stage int

const (
	configStage stage = iota
	timerStage
)

const (
	focusName = iota
	focusDuration
	breakName
	breakDuration
	tags
)

func getCursorColor() color.Color {
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	lightDark := lipgloss.LightDark(isDark)
	cursorColor := lightDark(lipgloss.Color("#b84c30"), lipgloss.Color("#e08060"))

	return cursorColor
}

func newModel(database *db.DB) model {
	placeholders := []string{
		"Press Tab to use default name (Focus)",
		"Duration in minutes",
		"Press Tab for default (Break), Esc to skip",
		"Duration in minutes",
		"Comma-separated tags",
	}

	inputs := make([]textinput.Model, len(placeholders))
	cursorColor := getCursorColor()

	for i, placeholder := range placeholders {
		input := textinput.New()
		input.Placeholder = placeholder

		input.SetWidth(55)
		s := input.Styles()
		s.Cursor.Color = cursorColor
		s.Cursor.Blink = true
		s.Blurred.Text = s.Focused.Text
		s.Blurred.Prompt = s.Focused.Prompt
		input.SetStyles(s)

		inputs[i] = input
	}

	inputs[focusName].Focus()

	return model{
		stage:    configStage,
		database: database,
		inputs:   inputs,
	}
}

func (m model) configView() tea.View {
	var b strings.Builder
	b.WriteString("Configure your session\n\n")

	labels := []string{
		"Focus name",
		"Focus duration",
		"Break name",
		"Break duration",
		"Tags",
	}

	for i, input := range m.inputs {
		label := lipgloss.NewStyle().Width(18).Render(labels[i])
		fmt.Fprintf(&b, "%s %s\n\n", label, input.View())
	}

	if m.err != "" {
		fmt.Fprintf(&b, "%s\n\n", m.err)
	}

	b.WriteString("Enter: continue  |  Tab: default  |  Esc: skip break  |  Ctrl+C: quit")

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	v := tea.NewView(style.Render(b.String()))
	v.AltScreen = true

	return v
}
