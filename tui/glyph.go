package tui

import (
	"math"
	"strings"
)

const bigTextHeight = 5

var bigTextGlyphs = map[rune][bigTextHeight]string{
	'0': {"╭─╮", "│ │", "│ │", "│ │", "╰─╯"},
	'1': {" ╷ ", " │ ", " │ ", " │ ", " ╵ "},
	'2': {"╭─╮", "  │", "╭─╯", "│  ", "╰─╯"},
	'3': {"╭─╮", "  │", " ─╯", "  │", "╰─╯"},
	'4': {"  ╷", "│ │", "╰─╯", "  │", "  ╵"},
	'5': {"╭─╮", "│  ", "╰─╮", "  │", "╰─╯"},
	'6': {"╭─╮", "│  ", "├─╮", "│ │", "╰─╯"},
	'7': {"╭─╮", "  │", "  │", "  │", "  ╵"},
	'8': {"╭─╮", "│ │", "├─┤", "│ │", "╰─╯"},
	'9': {"╭─╮", "│ │", "╰─┤", "  │", "╰─╯"},
	':': {"   ", " ╷ ", "   ", " ╵ ", "   "},
}

func bigText(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}

	lines := make([]strings.Builder, bigTextHeight)
	for i := range lines {
		lines[i].Grow(len(runes) * 4)
	}

	for i, r := range runes {
		glyph, ok := bigTextGlyphs[r]
		if !ok {
			glyph = fallbackGlyph(r)
		}

		for row := range lines {
			if i > 0 {
				lines[row].WriteByte(' ')
			}
			lines[row].WriteString(glyph[row])
		}
	}

	rows := make([]string, bigTextHeight)
	for i := range lines {
		rows[i] = lines[i].String()
	}
	return strings.Join(rows, "\n")
}

const (
	cupInteriorWidth = 11
	cupInteriorRows  = 6
)

func coffeeCup(ratio float64) string {
	if math.IsNaN(ratio) {
		ratio = 0
	}
	ratio = math.Max(0, math.Min(1, ratio))

	filled := int(math.Ceil(ratio * cupInteriorRows))

	empty := strings.Repeat(" ", cupInteriorWidth)
	full := strings.Repeat("▓", cupInteriorWidth)

	rows := make([]string, 0, cupInteriorRows+4)
	if filled > 0 {
		rows = append(rows, "      ~ ~ ~      ")
	} else {
		rows = append(rows, "                 ")
	}

	for i := range cupInteriorRows {
		level := empty
		if i >= cupInteriorRows-filled {
			level = full
		}

		handle := "  "
		switch i {
		case 1:
			handle = "──┐"
		case 2, 3:
			handle = "  │"
		case 4:
			handle = "──┘"
		}

		rows = append(rows, "  │"+level+"│"+handle)
	}

	rows = append(rows, "  ╰"+strings.Repeat("─", cupInteriorWidth)+"╯  ")

	return strings.Join(rows, "\n")
}

func fallbackGlyph(r rune) [bigTextHeight]string {
	var glyph [bigTextHeight]string
	for i := range glyph {
		glyph[i] = "   "
	}
	glyph[bigTextHeight/2] = " " + string(r) + " "
	return glyph
}
