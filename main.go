package main

import (
	"fmt"
	"os"

	"github.com/jensgreen/dux/app"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
	"github.com/jensgreen/dux/treemap/tiling"
)

func main() {
	path, debug := app.ArgsOrExit()
	logging.Setup(debug)

	fileEvents := make(chan files.FileEvent)
	stateEvents := make(chan dux.StateEvent)
	commands := make(chan dux.Command)

	initState := dux.State{
		MaxDepth:       2,
		IsWalkingFiles: true,
	}

	pres := dux.NewPresenter(
		fileEvents,
		commands,
		stateEvents,
		initState,
		tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Right: 1, Bottom: 1, Left: 1}),
	)
	app := app.NewApp(path, stateEvents, commands)

	go pres.Loop()
	go files.WalkDir(path, fileEvents, os.ReadDir)
	err := app.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
