package app

import (
	"errors"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
)

// Application2 represents an event-driven application running on a screen.
type Application2 struct {
	widget      views.Widget
	screen      tcell.Screen
	err         error
	wg          sync.WaitGroup
	stateSetter func(dux.StateUpdate)

	stateChan <-chan dux.StateUpdate
}

func (app *Application2) Draw() {
	app.widget.Draw()
	app.screen.Show()
}

// SetRootWidget sets the primary (root, main) Widget to be displayed.
func (app *Application2) SetRootWidget(widget views.Widget) {
	app.widget = widget
}

// initialize initializes the application.  It will normally attempt to
// allocate a default screen if one is not already established.
func (app *Application2) initialize() error {
	if app.screen == nil {
		if app.screen, app.err = tcell.NewScreen(); app.err != nil {
			return app.err
		}
	}
	return nil
}

func (app *Application2) terminateEventLoop() {
	ev := &eventAppQuit{}
	ev.SetEventNow()
	if scr := app.screen; scr != nil {
		go func() { scr.PostEventWait(ev) }()
	}
}

// SetScreen sets the screen to use for the application.  This must be
// done before the application starts to run or is initialized.
func (app *Application2) SetScreen(scr tcell.Screen) {
	if app.screen == nil {
		app.screen = scr
		app.err = nil
	}
}

func (app *Application2) Suspend() {
	if app.screen != nil {
		app.clearAlternateScreen()
		app.screen.Suspend()
	}
}

func (app *Application2) Resume() {
	if app.screen != nil {
		app.screen.Resume()
	}
}

func (app *Application2) SetStateChan(ch <-chan dux.StateUpdate) {
	app.stateChan = ch
}

func (app *Application2) SetStateSetter(cb func(dux.StateUpdate)) {
	app.stateSetter = cb
}

// Draw loop
func (app *Application2) drawLoop() {
	defer app.wg.Done()
loop:
	for {
		event, ok := <-app.stateChan
		if ok {
			if event.State.Quit {
				// TODO we actually the last (and only the last) alternate screen
				// to end up in the scrollback buffer
				app.clearAlternateScreen()
				app.terminateEventLoop()
				break loop
			}
			// tv.spinner.Tick()
			app.stateSetter(event)
			if event.State.Refresh != nil {
				event.State.Refresh.Do(app.Refresh)
			}
			app.Draw()
		} else {
			// Channel is closed. Set to nil channel, which is never selected.
			// This will keep the app on-screen, waiting for user's quit signal.
			app.stateChan = nil
		}
	}
}

// Event loop
func (app *Application2) eventLoop() {
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
		case *eventAppQuit:
			break loop
		default:
			widget.HandleEvent(ev)
		}
	}
}

func (app *Application2) startEventLoop() {
	app.wg.Add(1)
	go app.eventLoop()
}

func (app *Application2) startDrawLoop() {
	app.wg.Add(1)
	go app.drawLoop()
}

func (app *Application2) Run() error {
	screen := app.screen
	widget := app.widget

	if widget == nil {
		return errors.New("root widget not set")
	}
	if screen == nil {
		if e := app.initialize(); e != nil {
			return e
		}
		screen = app.screen
	}
	defer func() {
		screen.Fini()
	}()
	screen.Init()
	screen.Clear()
	widget.SetView(screen)

	app.startEventLoop()
	app.startDrawLoop()
	app.wg.Wait()
	return app.err
}

func (app *Application2) Refresh() {
	app.screen.Sync()
}

func (app *Application2) clearAlternateScreen() {
	app.screen.Clear()
	app.screen.Show()
}

func NewApplication2(stateChan <-chan dux.StateUpdate) *Application2 {
	app := &Application2{}
	app.SetStateChan(stateChan)

	return app
}

type eventAppQuit struct {
	tcell.EventTime
}
