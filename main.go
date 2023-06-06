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
	stateEvents := make(chan dux.StateEvent)
	commands := make(chan dux.Command)

	initState := dux.State{}
	shutdownCtx, shutdownFunc := context.WithCancel(context.Background())

	pres := dux.NewPresenter(
		shutdownCtx,
		fileEvents,
		commands,
		stateEvents,
		initState,
		tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Right: 1, Bottom: 1, Left: 1}),
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
