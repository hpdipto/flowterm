package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) sessionView() tea.View {
	s := m.sessions[m.sessionIndex]
	content := fmt.Sprintf("%s\n%02d:%02d", s.name, s.minute, s.second)

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	v := tea.NewView(style.Render(content))
	v.AltScreen = true

	return v
}
