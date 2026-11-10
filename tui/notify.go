package tui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hpdipto/flowterm/utils"
)

// OSC 777 makes the terminal own the notification, so clicking focuses it.
func notify(title, body string) tea.Cmd {
	return func() tea.Msg {
		seq := "\x1b]777;notify;" + sanitize(title) + ";" + sanitize(body) + "\x07"
		if err := writeNotification(seq); err == nil {
			playSound()
			return nil
		}

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

		playSound()

		return nil
	}
}

// sanitize strips control bytes and ';' that would break the escape sequence.
func sanitize(text string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ';':
			return ','
		case r < 0x20 || r == 0x7f:
			return -1
		}

		return r
	}, text)
}

var writeNotification = writeTerminal

func writeTerminal(seq string) error {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer utils.LogDeferredError(tty.Close)

	_, err = tty.WriteString(seq)

	return err
}

func playSound() {
	_ = exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
}
