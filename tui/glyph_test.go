package tui

import (
	"strings"
	"testing"
)

func TestCoffeeCupFillLevels(t *testing.T) {
	fullBlocks := cupInteriorRows * cupInteriorWidth

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
			if got := strings.Count(coffeeCup(tc.ratio), "▓"); got != tc.want {
				t.Fatalf("blocks = %d, want %d", got, tc.want)
			}
		})
	}

	previous := -1
	for step := 0; step <= 10; step++ {
		ratio := float64(step) / 10
		rows := strings.Count(coffeeCup(ratio), "▓") / cupInteriorWidth

		if rows < previous {
			t.Fatalf("coffee rows fell from %d to %d as ratio rose to %.1f", previous, rows, ratio)
		}
		previous = rows
	}
}
