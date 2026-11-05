package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	tea "charm.land/bubbletea/v2"
)

func notify(title, body string) tea.Cmd {
	return func() tea.Msg {
		script := fmt.Sprintf(
			"display notification %s with title %s",
			strconv.Quote(body),
			strconv.Quote(title),
		)

		err := exec.Command("osascript", "-e", script).Run()
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Print("\a")
			return nil
		}
		_ = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()

		return nil
	}
}
