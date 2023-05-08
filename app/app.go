package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
)

type App struct {
	path   string
	width  int
	height int

	main        *MainPanel
	titleBar    *TitleBar
	treemapView *TreemapView

	screen tcell.Screen
	view   views.View
	widget views.Widget
	wg     sync.WaitGroup
	err    error

	fileEvents    chan files.FileEvent
	treemapEvents chan dux.StateUpdate
	commands      chan dux.Command
	errors        []error

	views.WidgetWatchers
}

func (a *App) SetState(ev dux.StateUpdate) {
	a.errors = ev.Errors
	a.titleBar.SetState(ev.State)
	a.treemapView.SetState(ev.State)
}

func (a *App) Draw() {
	a.printErrors()
	a.widget.Draw()
	a.screen.Show()
}

func (a *App) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		mod, key, ch := ev.Modifiers(), ev.Key(), ev.Rune()
		log.Printf("EventKey Modifiers: %d Key: %d Rune: %d", mod, key, ch)
		switch key {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				a.commands <- dux.Quit{}
				return true
			case '+':
				a.commands <- dux.IncreaseMaxDepth{}
				return true
			case '-':
				a.commands <- dux.DecreaseMaxDepth{}
				return true
			}
		case tcell.KeyEscape, tcell.KeyCtrlC:
			a.commands <- dux.Quit{}
			return true
		case tcell.KeyCtrlL:
			a.commands <- dux.Refresh{}
			return true
		}
	case *tcell.EventResize:
		w, h := ev.Size()
		a.width = w
		a.height = h
		a.Resize()
		return true
	}
	return a.widget.HandleEvent(ev)
}

func (a *App) Run() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	err := a.init()
	if err != nil {
		return err
	}
	defer func() {
		a.screen.Fini()
	}()

	a.widget.SetView(a.view)
	fileEvents := make(chan files.FileEvent)
	treemapEvents := make(chan dux.StateUpdate)
	commands := make(chan dux.Command)

	a.fileEvents = fileEvents
	a.treemapEvents = treemapEvents
	a.commands = commands

	go signalHandler(commands)
	go files.WalkDir(a.path, fileEvents, os.ReadDir)

	initState := dux.State{
		MaxDepth:       2,
		IsWalkingFiles: true,
	}

	pres := dux.NewPresenter(
		fileEvents,
		commands,
		treemapEvents,
		initState,
		dux.WithPadding(dux.SliceAndDice{}, 1.0),
	)
	go pres.Loop()

	a.startEventLoop()
	a.startDrawLoop()
	a.wg.Wait()
	return err
}
func (app *App) Refresh() {
	app.screen.Sync()
}

func (a *App) Resize() {
	a.view.Resize(0, 0, a.width, a.height)
	a.widget.Resize()
	titlebarHeight := 1
	a.commands <- dux.Resize{
		Width:  a.width,
		Height: a.height - titlebarHeight,
	}
	a.Refresh()
	a.PostEventWidgetResize(a)
}

func (app *App) Suspend() {
	if app.screen != nil {
		app.clearAlternateScreen()
		app.screen.Suspend()
	}
}

func (app *App) Resume() {
	if app.screen != nil {
		app.screen.Resume()
	}
}

func (a *App) SetView(view views.View) {
	a.view = view
	if a.widget != nil {
		a.widget.SetView(view)
	}
}

func (a *App) Size() (int, int) {
	return a.widget.Size()
}

// Draw loop
func (app *App) drawLoop() {
	defer app.wg.Done()
loop:
	for {
		event, ok := <-app.treemapEvents
		if ok {
			if event.State.Quit {
				// TODO we actually the last (and only the last) alternate screen
				// to end up in the scrollback buffer
				app.clearAlternateScreen()
				app.terminateEventLoop()
				break loop
			}
			app.SetState(event)
			if event.State.Refresh != nil {
				event.State.Refresh.Do(app.Refresh)
			}
			app.Draw()
		} else {
			// Channel is closed. Set to nil channel, which is never selected.
			// This will keep the app on-screen, waiting for user's quit signal.
			app.treemapEvents = nil
		}
	}
}

// Event loop
func (app *App) eventLoop() {
	defer app.wg.Done()
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

func (app *App) startDrawLoop() {
	app.wg.Add(1)
	go app.drawLoop()
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
func (a *App) init() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		a.err = err
		return err
	}

	err = screen.Init()
	if err != nil {
		a.err = err
		return err
	}

	screen.Clear()
	a.screen = screen
	a.SetView(screen)
	return nil
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
func (a *App) printErrors() {
	errs := a.errors
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

	a.Suspend()
	fmt.Fprintln(os.Stderr, lines)
	a.Resume()
}

func (app *App) clearAlternateScreen() {
	app.screen.Clear()
	app.screen.Show()
}

func NewApp(path string) *App {
	tv := NewTreemapView()
	title := NewTitleBar()
	main := NewMainPanel(title, tv)

	app := &App{
		path:        path,
		main:        main,
		widget:      main,
		titleBar:    title,
		treemapView: tv,
	}

	return app
}

type terminateEventLoopEvent struct {
	tcell.EventTime
}
