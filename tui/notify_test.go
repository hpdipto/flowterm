package tui

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestNotifyWritesOSC777(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	var got string
	restore := stubNotificationWriter(func(seq string) error {
		got = seq
		return nil
	})
	t.Cleanup(restore)

	notify("Focus complete", "Focus; session\x07 completed")()

	want := "\x1b]777;notify;Focus complete;Focus, session completed\x07"
	if got != want {
		t.Errorf("sequence = %q, want %q", got, want)
	}
}

func TestNotifyFallsBackToBell(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	restore := stubNotificationWriter(func(string) error {
		return errors.New("no terminal")
	})
	t.Cleanup(restore)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	original := os.Stdout
	os.Stdout = w

	notify("Focus complete", "Focus")()

	os.Stdout = original

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	if string(out) != "\a" {
		t.Errorf("output = %q, want bell", string(out))
	}
}

func stubNotificationWriter(stub func(string) error) func() {
	original := writeNotification
	writeNotification = stub

	return func() { writeNotification = original }
}
