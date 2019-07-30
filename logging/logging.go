package logging

import (
	"io"
	"log"
	"os"
)

func Setup(enable bool) {
	if enable {
		log.SetFlags(log.Lmicroseconds | log.Lshortfile)
		log.SetOutput(os.Stderr)
	} else {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}
}
