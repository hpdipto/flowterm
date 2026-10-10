package tui

import (
	"hpdipto/flowterm/config"
	"hpdipto/flowterm/db"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func parseTags(value string) []string {
	var result []string

	for _, tag := range strings.Split(value, ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			result = append(result, tag)
		}
	}

	return result
}

func (m model) startTimer() (tea.Model, tea.Cmd) {
	focusMinute, err := strconv.Atoi(strings.TrimSpace(m.inputs[focusDuration].Value()))
	if err != nil || focusMinute <= 0 {
		m.currentInput = focusDuration
		m.err = "Enter a positive focus duration."
		return m.goToInput(focusDuration)
	}

	breakMinute, _ := strconv.Atoi(strings.TrimSpace(m.inputs[breakDuration].Value()))

	c := config.Config{
		FocusSessionName: strings.TrimSpace(m.inputs[focusName].Value()),
		FocusMinute:      focusMinute,
		BreakSessionName: strings.TrimSpace(m.inputs[breakName].Value()),
		BreakMinute:      breakMinute,
		Tags:             parseTags(m.inputs[tags].Value()),
	}

	if c.FocusSessionName == "" {
		c.FocusSessionName = "Focus"
	}

	if c.BreakSessionName == "" {
		c.BreakMinute = 0
	}

	s := &db.Session{
		FocusSessionName:     c.FocusSessionName,
		BreakSessionName:     c.BreakSessionName,
		FocusSessionDuration: c.FocusMinute,
		BreakSessionDuration: c.BreakMinute,
		Interrupted:          false,
	}

	if err := m.database.CreateSession(s); err != nil {
		m.err = err.Error()
		return m, nil
	}

	for _, name := range c.Tags {
		tag, err := m.database.GetOrCreateTag(name)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}

		if err := m.database.AddTagToSession(s.ID, tag.ID); err != nil {
			m.err = err.Error()
			return m, nil
		}
	}

	m.sessions = []session{
		{s.ID, c.FocusSessionName, c.FocusMinute, 0},
	}

	if c.BreakMinute > 0 {
		m.sessions = append(m.sessions, session{
			s.ID,
			c.BreakSessionName,
			c.BreakMinute,
			0,
		})
	}

	m.stage = timerStage
	m.sessionIndex = 0

	return m, tick()
}
