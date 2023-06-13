package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/nav"
)

type App struct {
	path      string
	prevState *dux.State

	main     *views.Panel
	titleBar *TitleBar
	treemap  *TreemapWidget

	screen tcell.Screen
	view   views.View
	widget views.Widget
	wg     sync.WaitGroup

	stateEvents <-chan dux.StateEvent
	commands    chan<- dux.Command
}

func (app *App) Run() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	err := app.init()
	if err != nil {
		return err
	}
	defer app.cleanup()

	go signalHandler(app.commands)
	app.startEventLoop()
	app.startUpdateLoop()
	app.wg.Wait()
	return err
}

func (app *App) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		return app.handleKey(ev)
	case *tcell.EventMouse:
		return app.handleMouse(ev)
	case *tcell.EventResize:
		w, h := ev.Size()
		_, th := app.titleBar.Size()
		app.commands <- dux.Resize{
			AppSize:     z2.Point{X: w, Y: h},
			TreemapSize: z2.Point{X: w, Y: h - th},
		}
		return true
	}
	return app.widget.HandleEvent(ev)
}

func (app *App) handleKey(ev *tcell.EventKey) bool {
	mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
	log.Printf("EventKey Modifiers: %d Key: %d Rune: %d", mod, key, ch)
	switch key {
	case tcell.KeyRune:
		switch ev.Rune() {
		// navigation
		case 'h':
			app.commands <- dux.Navigate{Direction: nav.DirectionLeft}
		case 'j':
			app.commands <- dux.Navigate{Direction: nav.DirectionDown}
		case 'k':
			app.commands <- dux.Navigate{Direction: nav.DirectionUp}
		case 'l':
			app.commands <- dux.Navigate{Direction: nav.DirectionRight}
		// depth
		case '+':
			app.commands <- dux.IncreaseMaxDepth{}
		case '-':
			app.commands <- dux.DecreaseMaxDepth{}
		// zoom
		case 'i':
			app.commands <- dux.ZoomIn{}
		case 'o':
			app.commands <- dux.ZoomOut{}
		// misc
		case ' ':
			app.commands <- dux.TogglePause{}
		case 'q':
			app.commands <- dux.Quit{}
		}
	// alt. navigation
	case tcell.KeyLeft:
		app.commands <- dux.Navigate{Direction: nav.DirectionLeft}
	case tcell.KeyRight:
		app.commands <- dux.Navigate{Direction: nav.DirectionRight}
	case tcell.KeyUp:
		app.commands <- dux.Navigate{Direction: nav.DirectionUp}
	case tcell.KeyDown:
		app.commands <- dux.Navigate{Direction: nav.DirectionDown}
	case tcell.KeyEnter:
		app.commands <- dux.Navigate{Direction: nav.DirectionIn}
	case tcell.KeyBackspace2:
		app.commands <- dux.Navigate{Direction: nav.DirectionOut}
	// misc
	case tcell.KeyEscape, tcell.KeyCtrlC:
		app.commands <- dux.Quit{}
	case tcell.KeyCtrlL:
		app.commands <- dux.Refresh{}
	case tcell.KeyCtrlZ:
		app.commands <- dux.SendToBackground{}
	}
	// key events are always consumed here at the top level
	return true
}

func (app *App) handleMouse(ev *tcell.EventMouse) bool {
	mx, my := ev.Position()
	log.Printf("EventMouse Buttons: %#b Modifiers: %#b Position: (%d, %d)", ev.Buttons(), ev.Modifiers(), mx, my)

	isClick := ev.Buttons()&tcell.ButtonPrimary == tcell.ButtonPrimary
	if isClick {
		log.Printf("Mouse click!")

		if app.titleBar.HandleEvent(ev) {
			return true
		}
		_, titlebarHeight := app.titleBar.Size()
		ev := NewEventMouseLocal(mx, my-titlebarHeight, ev)
		_ = app.treemap.HandleEvent(ev)
	}
	return true
}

func (app *App) SetState(state dux.State) {
	app.titleBar.SetState(state)
	app.treemap.SetState(state)
}

func (app *App) Draw() {
	app.widget.Draw()
	app.screen.Show()
}

func (app *App) refresh() {
	app.screen.Sync()
}

func (app *App) resize(width int, height int) {
	app.view.Resize(0, 0, width, height)
	app.widget.Resize()
	app.refresh()
}

func (app *App) Suspend() {
	app.clearAlternateScreen()
	app.screen.Suspend()
}

func (app *App) Resume() {
	app.screen.Resume()
}

func (app *App) SetView(view views.View) {
	app.view = view
	app.widget.SetView(view)
}

func (app *App) Size() (int, int) {
	return app.widget.Size()
}

// updateLoop polls for state updates and the quit signal, both sent by the presenter
func (app *App) updateLoop() {
	defer app.wg.Done()
	defer app.cleanupAndRepanic()
loop:
	for {
		event := <-app.stateEvents
		if event.State.Quit {
			// TODO we actually the last (and only the last) alternate screen
			// to end up in the scrollback buffer
			app.clearAlternateScreen()
			app.terminateEventLoop()
			break loop
		}
		app.printErrors(event.Errors)

		if app.prevState != nil && app.prevState.AppSize != event.State.AppSize {
			app.resize(event.State.AppSize.X, event.State.AppSize.Y)
		}

		if event.State.Treemap != nil {
			app.SetState(event.State)
			app.Draw()
		}

		if event.State.Refresh != nil {
			event.State.Refresh.Do(app.refresh)
		}
		if event.State.SendToBackground != nil {
			event.State.SendToBackground.Do(app.sendToBackground)
		}
	}
}

// eventLoop polls for and handles tcell events. This may result in commands being sent to the presenter,
// but no direct state updates.
func (app *App) eventLoop() {
	defer app.wg.Done()
	defer app.cleanupAndRepanic()
	screen := app.screen
loop:
	for {
		var widget views.Widget
		if widget = app.widget; widget == nil {
			break
		}

		ev := screen.PollEvent()
		switch ev.(type) {
		case *terminateEventLoopEvent:
			break loop
		default:
			app.HandleEvent(ev)
		}
	}
}

func (app *App) startEventLoop() {
	app.wg.Add(1)
	go app.eventLoop()
}

func (app *App) startUpdateLoop() {
	app.wg.Add(1)
	go app.updateLoop()
}

func (app *App) terminateEventLoop() {
	ev := &terminateEventLoopEvent{}
	ev.SetEventNow()
	if scr := app.screen; scr != nil {
		go func() { scr.PostEventWait(ev) }()
	}
}

// initialize initializes the application.  It will normally attempt to
// allocate a default screen if one is not already established.
func (app *App) init() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	err = screen.Init()
	if err != nil {
		return err
	}

	screen.EnableMouse(tcell.MouseButtonEvents)
	screen.Clear()
	app.screen = screen
	app.SetView(screen)
	return nil
}

func (app *App) cleanup() {
	if app.screen != nil {
		app.screen.Fini()
	}
}

// cleanupAndRepanic restores state of the terminal, and then just repanics to
// kill all other goroutines.  If another goroutine panics, the terminal will
// still be messed up, so a cleaner shutdown on panic would be desirable.
func (app *App) cleanupAndRepanic() {
	if e := recover(); e != nil {
		app.cleanup()
		panic(e)
	}
}

// printErrors prints error messages to stderr in the Normal Screen Buffer. The
// error messages will be visible when the application is closed or
// backgrounded, as well as during a brief flicker which occurs when the
// Alternate Screen Buffer is disabled to allow writing to the normal screen
// buffer.
//
// For more info, see:
// - https://stackoverflow.com/questions/39188508/how-curses-preserves-screen-contents
// - https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h2-The-Alternate-Screen-Buffer
// - https://invisible-island.net/xterm/xterm.faq.html#xterm_tite
func (app *App) printErrors(errs []error) {
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

	app.Suspend()
	fmt.Fprintln(os.Stderr, lines)
	app.Resume()
}

func (app *App) clearAlternateScreen() {
	app.screen.Clear()
	app.screen.Show()
}

func (app *App) sendToBackground() {
	app.Suspend()
	sig := syscall.SIGSTOP
	pid := os.Getpid()
	err := syscall.Kill(pid, sig)
	if err != nil {
		log.Printf("could not send %s to pid %d", sig, pid)
	}
	app.Resume()
	app.Draw()
}

func NewApp(path string, stateEvents <-chan dux.StateEvent, commands chan<- dux.Command) *App {
	tv := NewTreemapWidget(commands)
	title := NewTitleBar(commands)

	main := &views.Panel{}
	main.SetTitle(title)
	main.SetContent(tv)

	app := &App{
		path:        path,
		main:        main,
		widget:      main,
		titleBar:    title,
		treemap:     tv,
		stateEvents: stateEvents,
		commands:    commands,
	}

	return app
}

type terminateEventLoopEvent struct {
	tcell.EventTime
}
