package tui

import (
	"strings"
	"testing"
)

func TestCoffeeCupFillLevels(t *testing.T) {
	const width, rows = 11, 6
	fullBlocks := rows * width

	cases := []struct {
		name  string
		ratio float64
		want  int
	}{
		{"full", 1, fullBlocks},
		{"empty", 0, 0},
		{"above one clamps to full", 2, fullBlocks},
		{"below zero clamps to empty", -1, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := strings.Count(coffeeCup(tc.ratio, width, rows), "▓"); got != tc.want {
				t.Fatalf("blocks = %d, want %d", got, tc.want)
			}
		})
	}

	previous := -1
	for step := range 11 {
		ratio := float64(step) / 10
		filledRows := strings.Count(coffeeCup(ratio, width, rows), "▓") / width

		if filledRows < previous {
			t.Fatalf("coffee rows fell from %d to %d as ratio rose to %.1f", previous, filledRows, ratio)
		}
		previous = filledRows
	}
}

func TestCoffeeCupGeometryClampsAndAligns(t *testing.T) {
	cases := []struct{ width, rows int }{
		{0, 0},
		{cupMinInteriorWidth, cupMinInteriorRows},
		{11, 6},
		{cupMaxInteriorWidth + 10, cupMaxInteriorRows + 10},
	}

	for _, tc := range cases {
		width := clampInt(tc.width, cupMinInteriorWidth, cupMaxInteriorWidth)
		rows := clampInt(tc.rows, cupMinInteriorRows, cupMaxInteriorRows)
		lines := strings.Split(coffeeCup(1, tc.width, tc.rows), "\n")

		if len(lines) != rows+2 {
			t.Fatalf("lines = %d, want %d for %dx%d", len(lines), rows+2, tc.width, tc.rows)
		}

		want := width + 7
		for _, line := range lines {
			if got := len([]rune(line)); got != want {
				t.Fatalf("line width = %d, want %d for %dx%d: %q", got, want, tc.width, tc.rows, line)
			}
		}
	}
}
