package main

import (
	"fmt"
	"os"

	"github.com/jensgreen/dux/app"
	"github.com/jensgreen/dux/logging"
)

func main() {
	path, debug := app.ArgsOrExit()
	logging.Setup(debug)

	app := app.NewApp(path)
	err := app.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
