package main

import (
	"fmt"
	"hpdipto/flowterm/utils"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type session struct {
	id     int64
	name   string
	minute int
	second int
}

type model struct {
	height       int
	width        int
	sessions     []session
	sessionIndex int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initModel(sessions []session) model {
	return model{
		sessions:     sessions,
		sessionIndex: 0,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	index := m.sessionIndex

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		session := &m.sessions[index]

		if session.second == 0 && session.minute == 0 {
			m.sessionIndex++

			if m.sessionIndex >= len(m.sessions) {
				utils.LogError(closeSession(database, session.id, false))
				return m, tea.Quit
			}

			return m, tick()
		}

		if session.second > 0 {
			session.second--
		} else {
			session.second = 59
			session.minute--
		}

		return m, tick()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			utils.LogError(closeSession(database, m.sessions[index].id, true))
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	index := m.sessionIndex
	if m.sessionIndex >= len(m.sessions) {
		return tea.NewView("")
	}

	s := fmt.Sprintf("%s\n%02d:%02d", m.sessions[index].name, m.sessions[index].minute, m.sessions[index].second)

	v := tea.NewView(style.Render(s))
	v.AltScreen = true

	return v
}

func InitTUi(sessionId int64, c config) {
	sessions := []session{
		{sessionId, c.focusSessionName, c.focusMinute, 0},
	}

	if c.breakMinute > 0 {
		sessions = append(sessions, session{sessionId, c.breakSessionName, c.breakMinute, 0})
	}

	p := tea.NewProgram(initModel(sessions))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
