package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jensgreen/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
)

func printScreen(s tcell.SimulationScreen) {
	cells, w, _ := s.GetContents()
	var b strings.Builder
	for i, cell := range cells {
		for _, char := range cell.Runes {
			b.WriteRune(char)
		}
		if (i+1)%w == 0 {
			b.WriteRune('\n')
		}
	}
	fmt.Print(b.String())
}

func main() {
	path, debug := dux.ArgsOrExit()
	logging.Setup(debug)
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)
	s := tcell.NewSimulationScreen("utf-8")
	err := s.Init()
	scale := 1
	s.SetSize(80*scale, 24*scale)
	s.Clear()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer s.Fini()
	s.Clear()

	fileEvents := make(chan files.FileEvent)
	go func() {
		// walk all files, then quit
		files.WalkDir(path, fileEvents, os.ReadDir)
		s.InjectKey(tcell.KeyRune, 'q', tcell.ModNone)
	}()
	treemapEvents := make(chan dux.StateUpdate)
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

	// Set initial size.
	// tcell.SimulationScreen doesn't send EventResize on Init() like the normal tcell.Screen.
	commands <- dux.Resize{WindowWidth: 80 * scale, WindowHeight: 24 * scale}

	tui.MainLoop()
	printScreen(s)
}
