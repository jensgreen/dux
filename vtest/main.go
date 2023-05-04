package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

func NewTitleBar() *views.TextBar {
	tb := views.NewTextBar()
	style := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)

	tb.SetStyle(style)
	tb.SetLeft("title left", style)
	tb.SetCenter("dux (center)", style)
	tb.SetRight("(right)", style)

	return tb
}

type MainPanel struct {
	views.Panel
}

func NewMainPanel() *MainPanel {
	m := &MainPanel{}

	m.SetTitle(NewTitleBar())

	menu := views.NewText()
	menu.SetText("menu")
	m.SetMenu(menu)

	sb := views.NewText()
	sb.SetText("statusbar")
	m.SetStatus(sb)

	content := views.NewText()
	content.SetText("hello world")
	m.SetContent(content)

	return m
}

type App struct {
	app   *views.Application
	panel views.Widget
	view  views.View
	// main  *app.MainPanel

	views.WidgetWatchers
}

func (a *App) Draw() {
	a.panel.Draw()
}

func (a *App) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		// Intercept a few control keys up front, for global handling.
		case tcell.KeyCtrlC:
			a.app.Quit()
			return true
		case tcell.KeyCtrlL:
			a.app.Refresh()
			return true
		}
	}

	if a.panel != nil {
		return a.panel.HandleEvent(ev)
	}
	return false
}

func (a *App) Resize() {
	a.panel.Resize()
}

func (a *App) Run() error {
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	a.app.SetRootWidget(a)
	a.show(a.panel)

	a.app.Run()
	return nil
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

func NewApp() *App {
	app := &App{}
	app.app = &views.Application{}
	// main := NewMainPanel()
	// app.panel = main
	main := NewMainPanel()
	app.panel = main

	return app
}

func main() {
	a := NewApp()

	a.Run()
}
