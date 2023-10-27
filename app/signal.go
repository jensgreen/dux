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
	signals := make(chan os.Signal, 10) // must be buffered, or signal can be lost
	// SIGSTOP (stop) cannot be caught
	// SIGTSTP (stop, via ^Z) gets a default handler by the Go runtime
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		sig := <-signals
		log.Printf("Got signal %s", sig.String())
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Printf("signal: calling shutdown func")
			shutdown()
		default:
			// Whoops, signal.Notify() used, but no matching case
			log.Panicf("signal: got unhandled signal %s", sig)
		}
	}
}
