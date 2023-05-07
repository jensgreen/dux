package app

import (
	"errors"
	"log"
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
	stateSetter func(dux.State)

	stateChan   <-chan dux.StateUpdate
	commandChan chan<- dux.Command
}

func (app *Application2) Draw() {
	app.widget.Draw()
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

// // Refresh causes the application forcibly redraw everything.  Use this
// // to clear up screen corruption, etc.
// func (app *Application2) Refresh() {
// 	ev := &eventAppRefresh{}
// 	ev.SetEventNow()
// 	if scr := app.screen; scr != nil {
// 		go func() { scr.PostEventWait(ev) }()
// 	}
// }

// // Update asks the application to draw any screen updates that have not
// // been drawn yet.
// func (app *Application2) Update() {
// 	ev := &eventAppUpdate{}
// 	ev.SetEventNow()
// 	if scr := app.screen; scr != nil {
// 		go func() { scr.PostEventWait(ev) }()
// 	}
// }

// // PostFunc posts a function to be executed in the context of the
// // application's event loop.  Functions that need to update displayed
// // state, etc. can do this to avoid holding locks.
// func (app *Application2) PostFunc(fn func()) {
// 	ev := &eventAppFunc{fn: fn}
// 	ev.SetEventNow()
// 	if scr := app.screen; scr != nil {
// 		go func() { scr.PostEventWait(ev) }()
// 	}
// }

// SetScreen sets the screen to use for the application.  This must be
// done before the application starts to run or is initialized.
func (app *Application2) SetScreen(scr tcell.Screen) {
	if app.screen == nil {
		app.screen = scr
		app.err = nil
	}
}

func (app *Application2) SetStateChan(ch <-chan dux.StateUpdate) {
	app.stateChan = ch
}

func (app *Application2) SetCommandChan(ch chan<- dux.Command) {
	app.commandChan = ch
}

func (app *Application2) SetStateSetter(cb func(dux.State)) {
	app.stateSetter = cb
}

// Draw loop
func (app *Application2) drawLoop() {
	defer app.wg.Done()

	i := 0
loop:
	for {
		log.Printf("draw loop: %d", i)
		i++
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
			// tv.printErrors(event.Errors)
			app.stateSetter(event.State)
			event.State.Refresh.Do(app.refresh)
			app.Draw()
		} else {
			// Channel is closed. Set to nil channel, which is never selected.
			// This will keep the app on-screen, waiting for user's quit signal.
			app.stateChan = nil
		}
	}
	log.Printf("Leaving draw loop")
}

// Event loop
func (app *Application2) eventLoop() {
	defer app.wg.Done()

	screen := app.screen
	widget := app.widget

	i := 0
loop:
	for {
		log.Printf("event loop: %d", i)
		i++
		if widget = app.widget; widget == nil {
			break
		}

		ev := screen.PollEvent()
		switch nev := ev.(type) {
		case *eventAppQuit:
			break loop
		// case *eventAppUpdate:
		// 	panic("update event")
		// 	// screen.Show()
		// case *eventAppRefresh:
		// 	panic("refresh event")
		// 	// screen.Sync()
		// case *eventAppFunc:
		// 	nev.fn()
		case *tcell.EventResize:
			w, h := nev.Size()
			// TODO: command accepts treemap size, but this is window size
			app.commandChan <- dux.Resize{Width: w, Height: h}
		default:
			widget.HandleEvent(ev)
		}
	}
	log.Printf("Leaving event loop")
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
		return errors.New("Root widget not set")
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
	log.Printf("Application loops started")
	app.wg.Wait()
	log.Printf("Loops done, returning")
	return app.err
}

func (app *Application2) refresh() {
	app.screen.Sync()
}

func (app *Application2) clearAlternateScreen() {
	app.screen.Clear()
	app.screen.Show()
}

func NewApplication2(stateChan <-chan dux.StateUpdate, commandChan chan<- dux.Command) *Application2 {
	app := &Application2{}
	app.SetStateChan(stateChan)
	app.SetCommandChan(commandChan)

	return app
}

// type eventAppUpdate struct {
// 	tcell.EventTime
// }

type eventAppQuit struct {
	tcell.EventTime
}

// type eventAppRefresh struct {
// 	tcell.EventTime
// }

// type eventAppFunc struct {
// 	tcell.EventTime
// 	fn func()
// }
