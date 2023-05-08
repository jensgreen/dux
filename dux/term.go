package dux

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/z2"
)

type TerminalView struct {
	stateUpdates <-chan StateUpdate
	commands     chan<- Command
	screen       tcell.Screen
	quit         chan struct{}
	// spinner      *spinner
}

func NewTerminalView(screen tcell.Screen, events <-chan StateUpdate, commands chan<- Command) TerminalView {
	return TerminalView{
		screen:       screen,
		stateUpdates: events,
		commands:     commands,
		quit:         make(chan struct{}),
		// spinner:      newSpinner(),
	}
}

func (tv *TerminalView) MainLoop() {
	for tv.poll() {
		// loop
	}
}

func (tv *TerminalView) poll() bool {
	event, ok := <-tv.stateUpdates
	if ok {
		if event.State.Quit {
			// TODO we actually the last (and only the last) alternate screen
			// to end up in the scrollback buffer
			tv.clearAlternateScreen()
			return false
		}
		// tv.spinner.Tick()
		tv.printErrors(event.Errors)
		tv.update(event.State)
	} else {
		// Channel is closed. Set to nil channel, which is never selected.
		// This will keep the app on-screen, waiting for user's quit signal.
		tv.stateUpdates = nil
	}
	return true
}

func (tv *TerminalView) UserInputLoop() {
	s := tv.screen
	for {
		select {
		case <-tv.quit:
			return
		default:
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
				log.Printf("EventKey Modifiers: %d Key: %d Rune: %d", mod, key, ch)
				switch key {
				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q':
						tv.commands <- Quit{}
						return
					case '+':
						tv.commands <- IncreaseMaxDepth{}
					case '-':
						tv.commands <- DecreaseMaxDepth{}
					}
				case tcell.KeyEscape, tcell.KeyCtrlC:
					tv.commands <- Quit{}
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
			// case *tcell.EventResize:
			// 	// An EventRize will be sent on tcell.Screen.Init(), so there is no
			// 	// need to set an initial size.
			// 	w, h := ev.Size()
			// 	tv.commands <- Resize{Width: w, Height: h}
			}
		}
	}
}

func (tv *TerminalView) treemapSpaceToScreenSpace(treemapPane z2.Rect, statusbar z2.Rect) z2.Rect {
	rect := treemapPane
	rect.Y.Lo += statusbar.Y.Length()
	rect.Y.Hi += statusbar.Y.Length()
	return rect
}

func (tv *TerminalView) closeHalfOpen(rect z2.Rect) z2.Rect {
	// Hi is exclusive (half-open range)
	if rect.X.Hi > rect.X.Lo {
		rect.X.Hi--
	}
	if rect.Y.Hi > rect.Y.Lo {
		rect.Y.Hi--
	}
	return rect
}

func (tv *TerminalView) drawBox(rect z2.Rect) {
	lo := rect.Lo()
	hi := rect.Hi()

	// Upper edge
	style := tcell.StyleDefault
	for x := lo.X + 1; x < hi.X; x++ {
		tv.screen.SetContent(x, lo.Y, tcell.RuneHLine, nil, style)
	}
	// Lower edge
	for x := lo.X + 1; x < hi.X; x++ {
		tv.screen.SetContent(x, hi.Y, tcell.RuneHLine, nil, style)
	}
	// Left edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		tv.screen.SetContent(lo.X, y, tcell.RuneVLine, nil, style)
	}
	// Right edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		tv.screen.SetContent(hi.X, y, tcell.RuneVLine, nil, style)
	}
	// Corners, clockwise from lower right
	tv.screen.SetContent(hi.X, hi.Y, tcell.RuneLRCorner, nil, style)
	tv.screen.SetContent(lo.X, hi.Y, tcell.RuneLLCorner, nil, style)
	tv.screen.SetContent(lo.X, lo.Y, tcell.RuneULCorner, nil, style)
	tv.screen.SetContent(hi.X, lo.Y, tcell.RuneURCorner, nil, style)
}

func (tv *TerminalView) drawTreemapPane(state State) {
	itm := snapRoundTreemap(state.Treemap)
	isRoot := true
	tv.drawTreemap(state, itm, isRoot)
}

func (tv *TerminalView) drawTreemap(state State, tm intTreemap, isRoot bool) {
	f := tm.File
	rect := tm.Rect
	log.Printf("Drawing %s at %v", f.Path, rect)
	tv.drawBox(rect)
	tv.drawLabel(rect, f, isRoot)

	for _, child := range tm.Children {
		tv.drawTreemap(state, child, false)
	}
}

func (tv *TerminalView) drawLabel(rect z2.Rect, f files.File, isRoot bool) {
	xrangeLabel := z2.Interval{
		Lo: rect.X.Lo + 1, // don't draw on left corner
		Hi: rect.X.Hi,
	}

	fname := f.Name()
	if isRoot {
		fname = f.Path
	}
	style := tcell.StyleDefault
	if f.IsDir {
		if !strings.HasSuffix(fname, "/") {
			// avoid showing for example "/" as "//"
			fname += "/"
		}
		style = style.Foreground(tcell.ColorBlue)
	}

	// apply styling to name part only
	xrangeName := xrangeLabel
	if len(fname) < xrangeName.Hi-xrangeName.Lo {
		xrangeName.Hi = xrangeName.Lo + len(fname)
	}
	// draw disk usage in renaming range, if possible
	xrangeSize := z2.Interval{
		Lo: xrangeName.Hi,
		Hi: xrangeLabel.Hi,
	}
	y := rect.Lo().Y
	tv.drawString(xrangeName, y, fname, nil, style) // different when dir
	humanSize := " " + files.HumanizeIEC(f.Size)
	if xrangeSize.Hi >= xrangeSize.Lo+len(humanSize) {
		tv.drawString(xrangeSize, y, humanSize, nil, tcell.StyleDefault)
	}
}

// drawString draws the provided string on row y and in cols given by the open interval xrange,
// truncating the string if necessary
func (tv *TerminalView) drawString(xrange z2.Interval, y int, s string, combc []rune, style tcell.Style) {
	i := 0
	for x := xrange.Lo; x < xrange.Hi; x++ {
		if i == len(s) {
			break
		}
		tv.screen.SetContent(x, y, rune(s[i]), combc, style)
		i++
	}
}

// func (tv *TerminalView) drawStatusbar(state State) {
// 	file := state.Treemap.File
// 	statusbar := state.StatusbarSpazeZ2
// 	var s string = fmt.Sprintf("%s %s (%d files)", file.Path, files.HumanizeIEC(file.Size), state.TotalFiles)
// 	if state.IsWalkingFiles {
// 		s = fmt.Sprintf("%s %s", s, tv.spinner.String())
// 	}
// 	s = fmt.Sprintf("%s (%d)", s, state.MaxDepth)
// 	s = fmt.Sprintf("%-*v", statusbar.X.Length()-1, s)

// 	y := statusbar.Y.Lo
// 	style := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
// 	tv.drawString(statusbar.X, y, s, nil, style)
// }

func (tv *TerminalView) update(state State) {
	tv.screen.Clear()
	// tv.drawStatusbar(state)
	tv.drawTreemapPane(state)
	tv.screen.Show()
}

// printErrors prints error messages to stderr in the Normal Screen Buffer. The
// error messages will be visible when the application is closed or
// backgrounded, as well as during a brief flicker which occurs when the
// Alternate Screen Buffer is disabled to allow writing to the normal screen
// buffer.
//
// For more info, see:
//
// - https://stackoverflow.com/questions/39188508/how-curses-preserves-screen-contents
//
// - https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-The-Alternate-Screen-Buffer
//
// - https://invisible-island.net/xterm/xterm.faq.html#xterm_tite
func (tv *TerminalView) printErrors(errs []error) {
	if len(errs) == 0 {
		return
	}

	var msgs []string = make([]string, len(errs))
	for i, err := range errs {
		var perr *fs.PathError
		if errors.As(err, &perr) {
			msgs[i] = fmt.Sprintf("cannot access '%s': %v", perr.Path, perr.Err)
		} else {
			msgs[i] = err.Error()
		}
	}
	lines := strings.Join(msgs, "\n")

	tv.clearAlternateScreen()
	tv.screen.Suspend()
	fmt.Fprintln(os.Stderr, lines)
	tv.screen.Resume()
}

// clearAlternateScreen prevents the alternate screen from scrolling into the
// terminal's scrollback buffer when we switch back to normal mode.
//
// On some terminal emulators (including iTerm2 with "save lines to scrollback
// in alternate screen mode" set and Windows Terminal) the alternate screen
// buffer is dumped to the scrollback buffer when switching to normal screen
// mode if the lines in the scrollback exceeds the window height. The effect is
// that the error messages we write to stderr in normal mode become interleaved
// between dumps of the alternate screen buffer, leaving the scrollback in a
// very noisy state upon exit.
func (tv *TerminalView) clearAlternateScreen() {
	tv.screen.Clear()
	tv.screen.Show()
}

// snapRoundTreemap rounds float coordinates in a Treemap to
// discrete coordinates in an IntTreemap
func snapRoundTreemap(tm Treemap) intTreemap {
	children := make([]intTreemap, len(tm.Children))
	for i, child := range tm.Children {
		children[i] = snapRoundTreemap(child)
	}

	return intTreemap{
		File:     tm.File,
		Rect:     z2.SnapRoundRect(tm.Rect),
		Children: children,
	}
}

type intTreemap struct {
	File     files.File
	Rect     z2.Rect
	Children []intTreemap
}
