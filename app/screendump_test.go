package app

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// sgr builds an ANSI SGR escape sequence from the given parameters.
func sgr(params string) string {
	return csi + params + sgrEnd
}

// sgrFg returns an SGR foreground color parameter for a palette index.
func sgrFg(idx int) string {
	if idx < colorStandardCount {
		return fmt.Sprintf(";%d", sgrFgBase+idx)
	}
	return fmt.Sprintf(";%d", sgrFgBase+sgrHighIntensityOffset+idx-colorStandardCount)
}

// sgrBg returns an SGR background color parameter for a palette index.
func sgrBg(idx int) string {
	if idx < colorStandardCount {
		return fmt.Sprintf(";%d", sgrBgBase+idx)
	}
	return fmt.Sprintf(";%d", sgrBgBase+sgrHighIntensityOffset+idx-colorStandardCount)
}

var (
	reset = sgr(sgrReset)

	// tcell color indices (from tcell's color constants):
	//   ColorGreen=2, ColorRed=9, ColorYellow=11, ColorBlue=12, ColorWhite=15
	fgRed    = sgrFg(9)
	fgYellow = sgrFg(11)
	fgBlue   = sgrFg(12)
	fgWhite  = sgrFg(15)
	bgGreen  = sgrBg(2)
	bgBlue   = sgrBg(12)
)

func initScreen(t *testing.T, w, h int) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(w, h)
	return screen
}

func TestCaptureScreen_PlainText(t *testing.T) {
	screen := initScreen(t, 10, 2)
	putString(screen, 0, 0, "hello", tcell.StyleDefault)
	putString(screen, 0, 1, "world", tcell.StyleDefault)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset)+"hello"+reset+"\n"+sgr(sgrReset)+"world"+reset, got)
}

func TestCaptureScreen_TrailingSpacesTrimmed(t *testing.T) {
	screen := initScreen(t, 10, 1)
	putString(screen, 0, 0, "hi", tcell.StyleDefault)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset)+"hi"+reset, got)
}

func TestCaptureScreen_TrailingSpacesWithColorPreserved(t *testing.T) {
	screen := initScreen(t, 5, 1)
	style := tcell.StyleDefault.Background(tcell.ColorBlue)
	for x := 0; x < 5; x++ {
		screen.SetContent(x, 0, ' ', nil, style)
	}
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+bgBlue)+"     "+reset, got)
}

func TestCaptureScreen_ForegroundColor(t *testing.T) {
	screen := initScreen(t, 5, 1)
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	putString(screen, 0, 0, "red", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+fgRed)+"red"+reset, got)
}

func TestCaptureScreen_BoldItalicYellow(t *testing.T) {
	screen := initScreen(t, 5, 1)
	style := tcell.StyleDefault.Bold(true).Italic(true).Foreground(tcell.ColorYellow)
	putString(screen, 0, 0, "hi", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+sgrBold+sgrItalic+fgYellow)+"hi"+reset, got)
}

func TestCaptureScreen_StyleChangeMidRow(t *testing.T) {
	screen := initScreen(t, 6, 1)
	putString(screen, 0, 0, "ab", tcell.StyleDefault.Foreground(tcell.ColorRed))
	putString(screen, 2, 0, "cd", tcell.StyleDefault.Foreground(tcell.ColorBlue))
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+fgRed)+"ab"+sgr(sgrReset+fgBlue)+"cd"+reset, got)
}

func TestCaptureScreen_BackgroundAndForeground(t *testing.T) {
	screen := initScreen(t, 3, 1)
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGreen)
	putString(screen, 0, 0, "bar", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+fgWhite+bgGreen)+"bar"+reset, got)
}

func TestCaptureScreen_StandardColor(t *testing.T) {
	screen := initScreen(t, 5, 1)
	// ColorBlack (index 0) is a standard color, should use base range (30-37).
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack)
	putString(screen, 0, 0, "dark", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+sgrFg(0))+"dark"+reset, got)
}

func TestCaptureScreen_HighIntensityColor(t *testing.T) {
	screen := initScreen(t, 5, 1)
	// ColorRed (index 9) is a high-intensity color, should use bright range (90-97).
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	putString(screen, 0, 0, "glow", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+sgrFg(9))+"glow"+reset, got)
}

func TestCaptureScreen_StandardAndHighIntensityBackground(t *testing.T) {
	screen := initScreen(t, 6, 1)
	// ColorGreen (index 2) is standard, ColorBlue (index 12) is high-intensity.
	putString(screen, 0, 0, "std", tcell.StyleDefault.Background(tcell.ColorGreen))
	putString(screen, 3, 0, "brt", tcell.StyleDefault.Background(tcell.ColorBlue))
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, sgr(sgrReset+sgrBg(2))+"std"+sgr(sgrReset+sgrBg(12))+"brt"+reset, got)
}

func TestCaptureScreen_EmptyScreen(t *testing.T) {
	screen := initScreen(t, 5, 2)
	screen.Show()

	got := CaptureScreen(screen)
	// All default spaces trimmed, just newline between rows.
	assert.Equal(t, "\n", got)
}

func TestCaptureScreen_ZeroSize(t *testing.T) {
	screen := initScreen(t, 0, 0)
	got := CaptureScreen(screen)
	assert.Equal(t, "", got)
}

func putString(screen tcell.SimulationScreen, x, y int, s string, style tcell.Style) {
	for _, r := range s {
		screen.SetContent(x, y, r, nil, style)
		x++
	}
}
