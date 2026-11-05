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
	cupMinInteriorWidth = 5
	cupMaxInteriorWidth = 21
	cupMinInteriorRows  = 3
	cupMaxInteriorRows  = 10
)

func clampInt(value, low, high int) int {
	return min(max(value, low), high)
}

func coffeeCup(ratio float64, width, rows int) string {
	width = clampInt(width, cupMinInteriorWidth, cupMaxInteriorWidth)
	rows = clampInt(rows, cupMinInteriorRows, cupMaxInteriorRows)

	if math.IsNaN(ratio) {
		ratio = 0
	}
	ratio = math.Max(0, math.Min(1, ratio))

	filled := int(math.Ceil(ratio * float64(rows)))

	empty := strings.Repeat(" ", width)
	full := strings.Repeat("▓", width)

	totalWidth := width + 7
	lines := make([]string, 0, rows+2)

	if filled > 0 {
		steam := "~ ~ ~"
		left := (totalWidth - len(steam)) / 2
		lines = append(lines, strings.Repeat(" ", left)+steam+strings.Repeat(" ", totalWidth-left-len(steam)))
	} else {
		lines = append(lines, strings.Repeat(" ", totalWidth))
	}

	handleHeight := max(3, rows*2/3)
	handleTop := (rows - handleHeight) / 2
	handleBottom := handleTop + handleHeight - 1

	for i := range rows {
		level := empty
		if i >= rows-filled {
			level = full
		}

		handle := "   "
		switch {
		case i == handleTop:
			handle = "──┐"
		case i == handleBottom:
			handle = "──┘"
		case i > handleTop && i < handleBottom:
			handle = "  │"
		}

		lines = append(lines, "  │"+level+"│"+handle)
	}

	lines = append(lines, "  ╰"+strings.Repeat("─", width)+"╯   ")

	return strings.Join(lines, "\n")
}

func fallbackGlyph(r rune) [bigTextHeight]string {
	var glyph [bigTextHeight]string
	for i := range glyph {
		glyph[i] = "   "
	}
	glyph[bigTextHeight/2] = " " + string(r) + " "
	return glyph
}
