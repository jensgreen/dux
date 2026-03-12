package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "\033[0mhello\033[0m\n\033[0mworld\033[0m", got)
}

func TestCaptureScreen_TrailingSpacesTrimmed(t *testing.T) {
	screen := initScreen(t, 10, 1)
	putString(screen, 0, 0, "hi", tcell.StyleDefault)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, "\033[0mhi\033[0m", got)
}

func TestCaptureScreen_TrailingSpacesWithColorPreserved(t *testing.T) {
	screen := initScreen(t, 5, 1)
	style := tcell.StyleDefault.Background(tcell.ColorBlue)
	for x := 0; x < 5; x++ {
		screen.SetContent(x, 0, ' ', nil, style)
	}
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, "\033[0;104m     \033[0m", got)
}

func TestCaptureScreen_ForegroundColor(t *testing.T) {
	screen := initScreen(t, 5, 1)
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	putString(screen, 0, 0, "red", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, "\033[0;91mred\033[0m", got)
}

func TestCaptureScreen_BoldItalic(t *testing.T) {
	screen := initScreen(t, 5, 1)
	style := tcell.StyleDefault.Bold(true).Italic(true).Foreground(tcell.ColorYellow)
	putString(screen, 0, 0, "hi", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, "\033[0;1;3;93mhi\033[0m", got)
}

func TestCaptureScreen_StyleChangeMidRow(t *testing.T) {
	screen := initScreen(t, 6, 1)
	putString(screen, 0, 0, "ab", tcell.StyleDefault.Foreground(tcell.ColorRed))
	putString(screen, 2, 0, "cd", tcell.StyleDefault.Foreground(tcell.ColorBlue))
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, "\033[0;91mab\033[0;94mcd\033[0m", got)
}

func TestCaptureScreen_BackgroundAndForeground(t *testing.T) {
	screen := initScreen(t, 3, 1)
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorGreen)
	putString(screen, 0, 0, "bar", style)
	screen.Show()

	got := CaptureScreen(screen)
	assert.Equal(t, "\033[0;97;42mbar\033[0m", got)
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
