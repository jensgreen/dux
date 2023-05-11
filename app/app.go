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

	main     *views.Panel
	titleBar *TitleBar
	treemap  *TreemapWidget

	screen tcell.Screen
	view   views.View
	widget views.Widget
	wg     sync.WaitGroup

	fileEvents  chan files.FileEvent
	stateEvents chan dux.StateEvent
	commands    chan dux.Command
}

func (app *App) Run() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	err := app.init()
	if err != nil {
		return err
	}
	defer func() {
		if app.screen != nil {
			app.screen.Fini()
		}
	}()

	go signalHandler(app.commands)
	go files.WalkDir(app.path, app.fileEvents, os.ReadDir)

	initState := dux.State{
		MaxDepth:       2,
		IsWalkingFiles: true,
	}

	pres := dux.NewPresenter(
		app.fileEvents,
		app.commands,
		app.stateEvents,
		initState,
		dux.WithPadding(dux.SliceAndDice{}, 1.0),
	)
	go pres.Loop()

	app.startEventLoop()
	app.startDrawLoop()
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
		app.width = w
		app.height = h
		app.Resize()
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
		case 'q':
			app.commands <- dux.Quit{}
			return true
		case '+':
			app.commands <- dux.IncreaseMaxDepth{}
			return true
		case '-':
			app.commands <- dux.DecreaseMaxDepth{}
			return true
		}
	case tcell.KeyEscape, tcell.KeyCtrlC:
		app.commands <- dux.Quit{}
		return true
	case tcell.KeyCtrlL:
		app.commands <- dux.Refresh{}
		return true
	}
	return true
}

func (app *App) handleMouse(ev *tcell.EventMouse) bool {
	mx, my := ev.Position()
	log.Printf("EventMouse Buttons: %#b Modifiers: %#b Pos: (%d, %d)", ev.Buttons(), ev.Modifiers(), mx, my)

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

func (app *App) Refresh() {
	app.screen.Sync()
}

func (app *App) Resize() {
	app.view.Resize(0, 0, app.width, app.height)
	app.widget.Resize()
	app.Refresh()
	_, th := app.titleBar.Size()
	app.commands <- dux.Resize{
		Width:  app.width,
		Height: app.height - th,
	}
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

// Draw loop
func (app *App) drawLoop() {
	defer app.wg.Done()
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
		if event.State.Selection != nil {
			log.Printf("Selection: %s", *event.State.Selection)
		}
		app.printErrors(event.Errors)
		app.SetState(event.State)
		if event.State.Refresh != nil {
			event.State.Refresh.Do(app.Refresh)
		}
		app.Draw()
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

func NewApp(path string) *App {
	fileEvents := make(chan files.FileEvent)
	stateEvents := make(chan dux.StateEvent)
	commands := make(chan dux.Command)

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
		fileEvents:  fileEvents,
		stateEvents: stateEvents,
		commands:    commands,
	}

	return app
}

type terminateEventLoopEvent struct {
	tcell.EventTime
}
