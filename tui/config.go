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
	uv "github.com/charmbracelet/ultraviolet"
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

func newConfigModel(database *db.DB) model {
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

	title := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		PaddingTop(2).
		Render("Configure your session")

	fields := lipgloss.NewStyle().
		Width(m.width).
		Height(max(m.height-5, 0)).
		Align(lipgloss.Center, lipgloss.Center).
		Render(b.String())

	footer := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render("Enter: continue  |  Tab: default  |  Shift+Tab: previous  |  Esc: skip break  |  Ctrl+C: quit")

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, title, fields, footer, ""))
	v.AltScreen = true

	return v
}

func (m model) configViewUpdate(msg uv.Event) (tea.Model, tea.Cmd) {
	input, cmd := m.inputs[m.currentInput].Update(msg)
	m.inputs[m.currentInput] = input
	return m, cmd
}

func (m model) configViewUpdateWithKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "tab":
		switch m.currentInput {
		case focusName:
			m.inputs[focusName].SetValue("Focus")
		case breakName:
			m.inputs[breakName].SetValue("Break")
		default:
			return m, nil
		}
		return m.nextInput()

	case "shift+tab":
		if m.currentInput > 0 {
			return m.goToInput(m.currentInput - 1)
		}
		return m, nil

	case "esc":
		if m.currentInput == breakName {
			m.inputs[breakName].SetValue("")
			m.inputs[breakDuration].SetValue("0")
			return m.goToInput(tags)
		}

	case "enter":
		return m.submitInput()
	}

	return m.configViewUpdate(msg)
}
