package app

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// CaptureScreen reads the contents of a tcell screen buffer and returns a
// string with ANSI escape sequences that reproduces the screen's appearance.
func CaptureScreen(screen tcell.Screen) string {
	w, h := screen.Size()
	if w == 0 || h == 0 {
		return ""
	}

	var buf strings.Builder
	var prevStyle tcell.Style
	styleWritten := false

	for y := 0; y < h; y++ {
		// Collect the row's cells so we can trim trailing default spaces.
		type cell struct {
			mainc rune
			combc []rune
			style tcell.Style
			width int
		}
		row := make([]cell, 0, w)
		for x := 0; x < w; x++ {
			mainc, combc, style, width := screen.GetContent(x, y)
			if width == 0 {
				// Continuation cell of a wide character; skip.
				continue
			}
			row = append(row, cell{mainc, combc, style, width})
		}

		// Trim trailing spaces that have the default style.
		trimmed := row
		for len(trimmed) > 0 {
			last := trimmed[len(trimmed)-1]
			if last.mainc == ' ' && len(last.combc) == 0 && last.style == tcell.StyleDefault {
				trimmed = trimmed[:len(trimmed)-1]
			} else {
				break
			}
		}

		for _, c := range trimmed {
			if !styleWritten || c.style != prevStyle {
				writeStyle(&buf, c.style)
				prevStyle = c.style
				styleWritten = true
			}
			if c.mainc == 0 {
				buf.WriteRune(' ')
			} else {
				buf.WriteRune(c.mainc)
			}
			for _, r := range c.combc {
				buf.WriteRune(r)
			}
		}

		// Reset style at end of line to prevent background bleed.
		if styleWritten {
			buf.WriteString("\033[0m")
			styleWritten = false
		}
		if y < h-1 {
			buf.WriteRune('\n')
		}
	}

	return buf.String()
}

// writeStyle emits an ANSI SGR sequence that sets the given style.
func writeStyle(buf *strings.Builder, style tcell.Style) {
	fg, bg, attrs := style.Decompose()

	buf.WriteString("\033[0")

	if attrs&tcell.AttrBold != 0 {
		buf.WriteString(";1")
	}
	if attrs&tcell.AttrDim != 0 {
		buf.WriteString(";2")
	}
	if attrs&tcell.AttrItalic != 0 {
		buf.WriteString(";3")
	}
	if attrs&tcell.AttrUnderline != 0 {
		buf.WriteString(";4")
	}
	if attrs&tcell.AttrBlink != 0 {
		buf.WriteString(";5")
	}
	if attrs&tcell.AttrReverse != 0 {
		buf.WriteString(";7")
	}
	if attrs&tcell.AttrStrikeThrough != 0 {
		buf.WriteString(";9")
	}

	writeColor(buf, fg, 30)
	writeColor(buf, bg, 40)

	buf.WriteRune('m')
}

// writeColor appends SGR parameters for a color. base is 30 for foreground
// or 40 for background.
func writeColor(buf *strings.Builder, c tcell.Color, base int) {
	if !c.Valid() {
		return
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		fmt.Fprintf(buf, ";%d;2;%d;%d;%d", base+8, r, g, b)
		return
	}
	idx := int(c & 0xff)
	switch {
	case idx < 8:
		fmt.Fprintf(buf, ";%d", base+idx)
	case idx < 16:
		fmt.Fprintf(buf, ";%d", base+60+idx-8)
	default:
		fmt.Fprintf(buf, ";%d;5;%d", base+8, idx)
	}
}
