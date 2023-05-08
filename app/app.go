package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
)

type App struct {
	path string

	app         *Application2
	main        *MainPanel
	panel       views.Widget
	titleBar    *TitleBar
	treemapView *TreemapView
	view        views.View

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
	a.panel.Draw()
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
		// views.Application.run() handles EventResize by delegating it to our Resize() method,
		// so no need to handle it here.
		panic("Unexpected EventResize")
	}

	return a.panel.HandleEvent(ev)
}

func (a *App) Resize() {
	a.panel.Resize()
	a.commands <- dux.Resize{
		Width:  a.treemapView.width,
		Height: a.treemapView.height,
	}
	a.PostEventWidgetResize(a)
}

func (a *App) Run() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	a.app.SetRootWidget(a)
	a.show(a.panel)
	fileEvents := make(chan files.FileEvent)
	treemapEvents := make(chan dux.StateUpdate)
	commands := make(chan dux.Command)

	a.fileEvents = fileEvents
	a.treemapEvents = treemapEvents
	a.commands = commands
	a.app.SetCommandChan(a.commands)
	a.app.SetStateChan(a.treemapEvents)
	a.app.SetStateSetter(a.SetState)

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

	err := a.app.Run()
	return err
}

func (a *App) SetView(view views.View) {
	a.view = view
	if a.panel != nil {
		a.panel.SetView(view)
	}
}

func (a *App) Size() (int, int) {
	return a.panel.Size()
}

func (a *App) show(w views.Widget) {
	if w != a.panel {
		a.panel.SetView(nil)
		a.panel = w
	}

	a.panel.SetView(a.view)
	// a.Resize()
	// a.app.Refresh()
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

	a.app.Suspend()
	fmt.Fprintln(os.Stderr, lines)
	a.app.Resume()
}

func NewApp(path string) *App {
	tv := NewTreemapView()
	title := NewTitleBar()
	main := NewMainPanel(title, tv)

	app := &App{
		app:         &Application2{},
		path:        path,
		main:        main,
		panel:       main,
		titleBar:    title,
		treemapView: tv,
	}

	return app
}

func run() error {
	path, debug := ArgsOrExit()
	logging.Setup(debug)
	disableTruecolor()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	err = s.Init()
	if err != nil {
		return err
	}
	defer s.Fini()
	s.Clear()

	fileEvents := make(chan files.FileEvent)
	treemapEvents := make(chan dux.StateUpdate)
	commands := make(chan dux.Command)

	go signalHandler(commands)
	go files.WalkDir(path, fileEvents, os.ReadDir)

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

	tui := dux.NewTerminalView(s, treemapEvents, commands)
	go tui.UserInputLoop()
	tui.MainLoop()
	return nil
}

func signalHandler(commands chan<- dux.Command) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signals
	log.Printf("Got signal %s", sig.String())
	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		commands <- dux.Quit{}
	}
}

// disableTruecolor makes us follow the terminal color scheme by disabling tcell's truecolor support
func disableTruecolor() error {
	return os.Setenv("TCELL_TRUECOLOR", "disable")
}
