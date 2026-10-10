package tui

import (
	"fmt"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/utils"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
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
	width        int
	height       int
	stage        stage
	database     *db.DB
	inputs       []textinput.Model
	currentInput int
	err          string
	sessions     []session
	sessionIndex int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		if m.stage != timerStage {
			return m, nil
		}

		if m.sessionIndex >= len(m.sessions) {
			return m, tea.Quit
		}

		s := &m.sessions[m.sessionIndex]

		if s.second == 0 && s.minute == 0 {
			m.sessionIndex++

			if m.sessionIndex >= len(m.sessions) {
				utils.LogError(db.CloseSession(m.database, s.id, false))
				return m, tea.Quit
			}

			return m, tick()
		}

		if s.second > 0 {
			s.second--
		} else {
			s.second = 59
			s.minute--
		}

		return m, tick()

	case tea.KeyPressMsg:
		if m.stage == timerStage {
			switch msg.String() {
			case "ctrl+c", "q":
				utils.LogError(db.CloseSession(
					m.database,
					m.sessions[m.sessionIndex].id,
					true,
				))
				return m, tea.Quit
			}

			return m, nil
		}

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

		case "esc":
			if m.currentInput == breakName {
				m.inputs[breakName].SetValue("")
				m.inputs[breakDuration].SetValue("0")
				return m.goToInput(tags)
			}

		case "enter":
			return m.submitInput()
		}
	}

	if m.stage == configStage {
		input, cmd := m.inputs[m.currentInput].Update(msg)
		m.inputs[m.currentInput] = input
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.stage == configStage {
		var b strings.Builder

		b.WriteString("Configure your session\n\n")

		labels := []string{
			"Focus session name",
			"Focus duration (minutes)",
			"Break session name",
			"Break duration (minutes)",
			"Tags",
		}

		for i, input := range m.inputs {
			fmt.Fprintf(&b, "%s\n%s\n\n", labels[i], input.View())
		}

		if m.err != "" {
			fmt.Fprintf(&b, "%s\n\n", m.err)
		}

		b.WriteString("Enter: continue  |  Tab: default  |  Esc: skip break  |  Ctrl+C: quit")

		v := tea.NewView(b.String())
		v.AltScreen = true
		return v
	}

	if m.sessionIndex >= len(m.sessions) {
		return tea.NewView("")
	}

	s := m.sessions[m.sessionIndex]
	content := fmt.Sprintf(
		"%s\n%02d:%02d",
		s.name,
		s.minute,
		s.second,
	)

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	v := tea.NewView(style.Render(content))
	v.AltScreen = true
	return v
}

func InitTUi(database *db.DB) {
	p := tea.NewProgram(newModel(database))

	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
	}
}
