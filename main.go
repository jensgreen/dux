package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jensgreen/dux/app"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/logging"
	"github.com/jensgreen/dux/recovery"
	"github.com/jensgreen/dux/treemap/tiling"
)

func main() {
	path, debug := app.ArgsOrExit()
	logging.Setup(debug)

	fileEvents := make(chan files.FileEvent)
	stateEvents := make(chan dux.StateEvent, 1)
	commands := make(chan dux.Command, 1)

	initState := dux.State{}
	shutdownCtx, shutdownFunc := context.WithCancel(context.Background())
	go app.SignalHandler(commands, shutdownFunc)

	pres := dux.NewPresenter(
		shutdownCtx,
		shutdownFunc,
		fileEvents,
		commands,
		stateEvents,
		initState,
		tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Right: 1, Bottom: 1, Left: 1}),
		files.NewFS(),
	)
	app := app.NewApp(shutdownCtx, path, stateEvents, commands)

	rec := recovery.New(shutdownFunc)
	rec.Go(pres.Loop)
	rec.Go(func() {
		files.WalkDir(shutdownCtx, path, fileEvents, os.ReadDir)
	})
	err := app.Run()
	rec.Release()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
