package tui

import (
	"fmt"
	"hpdipto/flowterm/activity"
	"hpdipto/flowterm/db"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type stage int

const (
	configStage stage = iota
	timerStage
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
	tracker      *tracker
}

type tickMsg time.Time

type sampleTickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func sampleTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return sampleTickMsg(t)
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
		return m, nil

	case tickMsg:
		return m.sessionViewUpdate()

	case sampleTickMsg:
		if m.tracker == nil {
			return m, nil
		}

		if s, err := activity.Current(); err == nil {
			phase := db.PhaseFocus
			if m.sessionIndex > 0 {
				phase = db.PhaseBreak
			}

			m.tracker.sample(time.Now(), s, phase)
		}

		return m, sampleTick()

	case tea.KeyPressMsg:
		if m.stage == timerStage {
			return m.sessionViewUpdateWithKey(msg)
		}
		return m.configViewUpdateWithKey(msg)
	}

	if m.stage == configStage {
		return m.configViewUpdate(msg)
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.stage == configStage {
		return m.configView()
	}

	if m.sessionIndex >= len(m.sessions) {
		return tea.NewView("")
	}

	return m.sessionView()
}

func InitTUi(database *db.DB) {
	p := tea.NewProgram(newConfigModel(database))

	if _, err := p.Run(); err != nil {
		fmt.Printf("TUI error: %v\n", err)
	}
}
