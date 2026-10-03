package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	minute int
	second int
	height int
	width  int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initModel(minute int) model {
	return model{
		minute: minute,
		second: 0,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.second -= 1
		if m.second < 0 {
			m.second = 59
			m.minute -= 1
		}
		if m.minute < 0 {
			return m, tea.Quit
		}
		return m, tick()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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

	s := fmt.Sprintf("Focus\n%02d:%02d", m.minute, m.second)

	v := tea.NewView(style.Render(s))
	v.AltScreen = true

	return v
}

func InitTUi(minute int) {
	p := tea.NewProgram(initModel(minute))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
