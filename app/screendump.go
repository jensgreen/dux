package app

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ANSI terminal escape constants.
// Reference: ECMA-48 §8.3.117, https://en.wikipedia.org/wiki/ANSI_escape_code#SGR
const (
	csi    = "\033[" // Control Sequence Introducer (ESC + '[')
	sgrEnd = "m"     // SGR (Select Graphic Rendition) sequence terminator

	sgrReset         = "0"
	sgrBold          = ";1"
	sgrDim           = ";2"
	sgrItalic        = ";3"
	sgrUnderline     = ";4"
	sgrBlink         = ";5"
	sgrReverse       = ";7"
	sgrStrikethrough = ";9"

	sgrFgBase = 30 // standard foreground colors: 30-37
	sgrBgBase = 40 // standard background colors: 40-47

	sgrColorExtendedOffset = 8  // base+8 = 38 (fg) or 48 (bg) for extended colors
	sgrColorRGB            = 2  // extended color sub-mode for 24-bit RGB
	sgrColor256            = 5  // extended color sub-mode for 256-color palette
	sgrHighIntensityOffset = 60 // offset from standard to bright colors (e.g., 30→90)

	colorIndexMask        = 0xff // mask to extract palette index from tcell.Color
	colorStandardCount    = 8    // standard colors: indices 0-7
	colorHighIntensityEnd = 16   // high-intensity colors: indices 8-15
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
			buf.WriteString(csi + sgrReset + sgrEnd)
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

	buf.WriteString(csi + sgrReset)

	if attrs&tcell.AttrBold != 0 {
		buf.WriteString(sgrBold)
	}
	if attrs&tcell.AttrDim != 0 {
		buf.WriteString(sgrDim)
	}
	if attrs&tcell.AttrItalic != 0 {
		buf.WriteString(sgrItalic)
	}
	if attrs&tcell.AttrUnderline != 0 {
		buf.WriteString(sgrUnderline)
	}
	if attrs&tcell.AttrBlink != 0 {
		buf.WriteString(sgrBlink)
	}
	if attrs&tcell.AttrReverse != 0 {
		buf.WriteString(sgrReverse)
	}
	if attrs&tcell.AttrStrikeThrough != 0 {
		buf.WriteString(sgrStrikethrough)
	}

	writeColor(buf, fg, sgrFgBase)
	writeColor(buf, bg, sgrBgBase)

	buf.WriteString(sgrEnd)
}

// writeColor appends SGR parameters for a color. base is sgrFgBase for
// foreground or sgrBgBase for background.
func writeColor(buf *strings.Builder, c tcell.Color, base int) {
	if !c.Valid() {
		return
	}
	if c.IsRGB() {
		r, g, b := c.RGB()
		fmt.Fprintf(buf, ";%d;%d;%d;%d;%d", base+sgrColorExtendedOffset, sgrColorRGB, r, g, b)
		return
	}
	idx := int(c & colorIndexMask)
	switch {
	case idx < colorStandardCount:
		fmt.Fprintf(buf, ";%d", base+idx)
	case idx < colorHighIntensityEnd:
		fmt.Fprintf(buf, ";%d", base+sgrHighIntensityOffset+idx-colorStandardCount)
	default:
		fmt.Fprintf(buf, ";%d;%d;%d", base+sgrColorExtendedOffset, sgrColor256, idx)
	}
}
