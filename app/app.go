package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
)

func Run() error {
	path, debug := argsOrExit()
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
