package tui

import "strings"

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

func fallbackGlyph(r rune) [bigTextHeight]string {
	var glyph [bigTextHeight]string
	for i := range glyph {
		glyph[i] = "   "
	}
	glyph[bigTextHeight/2] = " " + string(r) + " "
	return glyph
}
