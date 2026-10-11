package tui

import (
	"hpdipto/flowterm/db"
	"image/color"
	"os"

	"charm.land/bubbles/v2/textinput"
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
