package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
)

type App struct {
	path string

	app         *views.Application
	main        *MainPanel
	panel       views.Widget
	titleBar    *TitleBar
	treemapView *TreemapView
	view        views.View

	fileEvents    chan files.FileEvent
	treemapEvents chan dux.StateUpdate
	commands      chan dux.Command

	views.WidgetWatchers
}

func (a *App) Draw() {
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
				a.app.Quit()
				return true
			case '+':
				a.commands <- dux.IncreaseMaxDepth{}
				return true
			case '-':
				a.commands <- dux.DecreaseMaxDepth{}
				return true
			}
		case tcell.KeyEscape, tcell.KeyCtrlC:
			a.app.Quit()
			return true
		case tcell.KeyCtrlL:
			a.app.Refresh()
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

	go func() {
		for update := range treemapEvents {
			a.app.PostFunc(func() {
				a.titleBar.Update(update.State)
				a.treemapView.Update(update.State)
			})
		}
	}()

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
	a.app.PostFunc(func() {
		if w != a.panel {
			a.panel.SetView(nil)
			a.panel = w
		}

		a.panel.SetView(a.view)
		a.Resize()
		a.app.Refresh()
	})
}

func NewApp(path string) *App {
	tv := NewTreemapView()
	title := NewTitleBar()
	main := NewMainPanel(title, tv)

	app := &App{
		app:         &views.Application{},
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
