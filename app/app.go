package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/cancellable"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/nav"
)

type App struct {
	ctx  context.Context
	path string

	main      *views.Panel
	titleBar  *TitleBar
	treemap   *TreemapWidget
	statusBar *StatusBar

	screen tcell.Screen
	view   views.View
	widget views.Widget

	stateEvents <-chan dux.StateEvent
	commands    chan<- dux.Command
	tcellEvents chan tcell.Event
}

func (app *App) Run() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	err := app.init()
	if err != nil {
		return err
	}
	defer app.cleanup()

	go app.screen.ChannelEvents(app.tcellEvents, app.ctx.Done())
	app.loop()
	return nil
}

func (app *App) handleTcellEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		return app.handleKey(ev)
	case *tcell.EventMouse:
		return app.handleMouse(ev)
	case *tcell.EventResize:
		w, h := ev.Size()
		_, th := app.titleBar.Size()

		var cmd dux.Command = dux.Resize{
			AppSize:     z2.Point{X: w, Y: h},
			TreemapSize: z2.Point{X: w, Y: h - th},
		}
		err := cancellable.Send(app.ctx, app.commands, cmd)
		if err != nil {
			log.Println("could not handle EventResize:", err.Error())
		}
		return true
	}
	return app.widget.HandleEvent(ev)
}

func (app *App) handleKey(ev *tcell.EventKey) bool {
	mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
	log.Printf("EventKey Modifiers: %d Key: %d Rune: %d", mod, key, ch)
	var cmd dux.Command
	switch key {
	case tcell.KeyRune:
		switch ev.Rune() {
		// navigation
		case 'h':
			cmd = dux.Navigate{Direction: nav.DirectionLeft}
		case 'j':
			cmd = dux.Navigate{Direction: nav.DirectionDown}
		case 'k':
			cmd = dux.Navigate{Direction: nav.DirectionUp}
		case 'l':
			cmd = dux.Navigate{Direction: nav.DirectionRight}
		// depth
		case '+':
			cmd = dux.IncreaseMaxDepth{}
		case '-':
			cmd = dux.DecreaseMaxDepth{}
		// zoom
		case 'i':
			cmd = dux.ZoomIn{}
		case 'o':
			cmd = dux.ZoomOut{}
		// misc
		case ' ':
			cmd = dux.TogglePause{}
		case 'q':
			cmd = dux.Quit{}
		}
	// alt. navigation
	case tcell.KeyLeft:
		cmd = dux.Navigate{Direction: nav.DirectionLeft}
	case tcell.KeyRight:
		cmd = dux.Navigate{Direction: nav.DirectionRight}
	case tcell.KeyUp:
		cmd = dux.Navigate{Direction: nav.DirectionUp}
	case tcell.KeyDown:
		cmd = dux.Navigate{Direction: nav.DirectionDown}
	case tcell.KeyEnter:
		cmd = dux.Navigate{Direction: nav.DirectionIn}
	case tcell.KeyBackspace2:
		cmd = dux.Navigate{Direction: nav.DirectionOut}
	// misc
	case tcell.KeyEscape, tcell.KeyCtrlC:
		cmd = dux.Quit{}
	case tcell.KeyCtrlL:
		cmd = dux.Refresh{}
	case tcell.KeyCtrlZ:
		cmd = dux.SendToBackground{}
	}

	if cmd != nil {
		err := cancellable.Send(app.ctx, app.commands, cmd)
		if err != nil {
			log.Printf("could not send command %T: %s", cmd, err.Error())
		}
	}
	// key events are always consumed here at the top level
	return true
}

func (app *App) handleMouse(ev *tcell.EventMouse) bool {
	mx, my := ev.Position()
	log.Printf("EventMouse Buttons: %#b Modifiers: %#b Position: (%d, %d)", ev.Buttons(), ev.Modifiers(), mx, my)

	isClick := ev.Buttons()&tcell.ButtonPrimary != 0
	if isClick {
		log.Printf("Mouse click!")

		if app.titleBar.HandleEvent(ev) {
			return true
		}
		_, titlebarHeight := app.titleBar.Size()
		_, tmHeight := app.treemap.Size()
		if app.statusBar.HandleEvent(NewEventMouseLocal(ev, 0, titlebarHeight+tmHeight)) {
			return true
		}
		ev := NewEventMouseLocal(ev, 0, titlebarHeight)
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
	if err := app.screen.Suspend(); err != nil {
		log.Panicf("could not suspend: %s", err)
	}
}

func (app *App) Resume() {
	if err := app.screen.Resume(); err != nil {
		log.Panicf("could not resume: %s", err)
	}
}

func (app *App) SetView(view views.View) {
	app.view = view
	app.widget.SetView(view)
}

func (app *App) Size() (int, int) {
	return app.widget.Size()
}

func (app *App) loop() {
	defer app.cleanupAndRepanic()
	for {
		select {
		case <-app.ctx.Done():
			return
		case event := <-app.tcellEvents:
			// Poll for and handle tcell events.
			// May result in commands being sent to the presenter, but no direct
			// state updates.
			switch event.(type) {
			case *terminateEventLoopEvent:
				return
			default:
				app.handleTcellEvent(event)
			}
		case event := <-app.stateEvents:
			// Poll for state updates and the quit signal, both sent by the
			// presenter.
			quit := app.handleStateEvent(event)
			if quit {
				return
			}
		}
	}
}

func (app *App) handleStateEvent(event dux.StateEvent) (quit bool) {
	if event.State.Quit {
		log.Printf("App got Quit event, terminating updateLoop!")
		// TODO we actually the last (and only the last) alternate screen
		// to end up in the scrollback buffer
		app.clearAlternateScreen()
		app.closeTcellEventChannel()
		return true
	}
	app.printErrors(event.Errors)

	if event.State.Treemap != nil {
		app.SetState(event.State)
		app.Draw()
	}

	switch action := event.Action; action {
	case dux.ActionNone:
		// noop
	case dux.ActionResize:
		app.resize(event.State.AppSize.X, event.State.AppSize.Y)
	case dux.ActionRefresh:
		app.refresh()
	case dux.ActionBackground:
		app.sendToBackground()
		app.resumeFromBackground()
	default:
		log.Panicf("unhandled action: %d", action)
	}

	return false
}

func (app *App) closeTcellEventChannel() {
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
	pid := os.Getpid()
	log.Printf("stopping pid %d", pid)
	err := syscall.Kill(pid, syscall.SIGTSTP)
	if err != nil {
		log.Printf("could stop pid %d", pid)
	} else {
		// Zzzz... suspended
		log.Printf("pid %d woke up", pid)
	}
}

func (app *App) resumeFromBackground() {
	log.Printf("resuming from background")
	app.Resume()
	app.Draw()
}

func NewApp(ctx context.Context, path string, stateEvents <-chan dux.StateEvent, commands chan<- dux.Command) *App {
	tv := NewTreemapWidget(commands)
	title := NewTitleBar(commands)
	status := NewStatusBar(commands)

	main := &views.Panel{}
	main.SetTitle(title)
	main.SetContent(tv)
	main.SetStatus(status)

	app := &App{
		path:        path,
		main:        main,
		widget:      main,
		titleBar:    title,
		statusBar:   status,
		treemap:     tv,
		stateEvents: stateEvents,
		commands:    commands,
		tcellEvents: make(chan tcell.Event),
		ctx:         ctx,
	}

	return app
}

type terminateEventLoopEvent struct {
	tcell.EventTime
}
