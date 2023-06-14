package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jensgreen/dux/dux"
)

func SignalHandler(commands chan<- dux.Command, shutdown context.CancelFunc) {
	signals := make(chan os.Signal, 20) // must be buffered
	// SIGSTOP (stop) cannot be caught
	// SIGTSTP (stop, via ^Z) gets a default handles by the Go runtime
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGCONT)
	for {
		sig := <-signals
		log.Printf("Got signal %s", sig.String())
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Printf("signal: calling shutdown func")
			shutdown()
			log.Printf("signal: sending Quit")
			commands <- dux.Quit{}
			log.Printf("signal: sent Quit")
		case syscall.SIGCONT:
			log.Printf("signal: sending WakeUp")
			commands <- dux.WakeUp{}
			log.Printf("signal: sent WakeUp")
		default:
			// Whoops, signal.Notify() used, but no matching case
			log.Panicf("signal: got unhandled signal %s", sig)
		}
	}
}
