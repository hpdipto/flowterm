package tui

import (
	"io"
	"os"
	"testing"
)

func TestNotifyFallsBackToBell(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

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
