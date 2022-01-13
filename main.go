package main

import (
	"fmt"
	"os"

	"github.com/jensgreen/dux/app"
)

func main() {
	err := app.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
