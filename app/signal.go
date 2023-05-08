package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jensgreen/dux/dux"
)

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
