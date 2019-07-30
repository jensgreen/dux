package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/jensgreen/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
)

func main() {
	path, debug := dux.ArgsOrExit()
	logging.Setup(debug)
	disableTruecolor()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	err = s.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer s.Fini()
	s.Clear()

	fileEvents := make(chan files.FileEvent)
	treemapEvents := make(chan dux.StateUpdate)
	go files.WalkDir(path, fileEvents, os.ReadDir)
	commands := make(chan dux.Command)

	initState := dux.State{
		MaxDepth: 2,
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
}

// disableTruecolor makes us follow the terminal color scheme by disabling tcell's truecolor support
func disableTruecolor() error {
	return os.Setenv("TCELL_TRUECOLOR", "disable")
}
