package tui

import (
	"hpdipto/flowterm/db"

	"charm.land/bubbles/v2/textinput"
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

func newModel(database *db.DB) model {
	placeholders := []string{
		"Press Tab to use default name (Focus)",
		"Duration in minutes",
		"Press Tab for default (Break), Esc to skip",
		"Duration in minutes",
		"Comma-separated tags",
	}

	inputs := make([]textinput.Model, len(placeholders))

	for i, placeholder := range placeholders {
		input := textinput.New()
		input.Placeholder = placeholder
		input.SetWidth(55)
		inputs[i] = input
	}

	inputs[focusName].Focus()

	return model{
		stage:    configStage,
		database: database,
		inputs:   inputs,
	}
}
