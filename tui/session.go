package tui

import (
	"fmt"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/utils"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) sessionView() tea.View {
	s := m.sessions[m.sessionIndex]
	content := fmt.Sprintf("%s\n%s", s.name, bigText(fmt.Sprintf("%02d:%02d", s.minute, s.second)))

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	v := tea.NewView(style.Render(content))
	v.AltScreen = true

	return v
}

func (m model) sessionViewUpdate() (tea.Model, tea.Cmd) {
	if m.stage != timerStage {
		return m, nil
	}

	if m.sessionIndex >= len(m.sessions) {
		return m, tea.Quit
	}

	s := &m.sessions[m.sessionIndex]

	if s.second == 0 && s.minute == 0 {
		title := "Focus complete"
		if m.sessionIndex > 0 {
			title = "Break over"
		}
		cmd := notify(title, s.name+" session completed")

		m.sessionIndex++

		if m.sessionIndex >= len(m.sessions) {
			m.saveActivities()
			utils.LogError(db.CloseSession(m.database, s.id, false))
			return m, tea.Sequence(cmd, tea.Quit)
		}

		return m, tea.Batch(cmd, tick())
	}

	if s.second > 0 {
		s.second--
	} else {
		s.second = 59
		s.minute--
	}

	return m, tick()
}

func (m model) sessionViewUpdateWithKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.sessionIndex >= len(m.sessions) {
			return m, tea.Quit
		}

		m.saveActivities()
		utils.LogError(db.CloseSession(
			m.database,
			m.sessions[m.sessionIndex].id,
			true,
		))
		return m, tea.Quit
	}

	return m, nil
}

func (m model) saveActivities() {
	if m.tracker == nil || len(m.sessions) == 0 {
		return
	}

	utils.LogError(m.tracker.stop())
}
