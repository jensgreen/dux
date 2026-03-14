package app

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func printUsage(w io.Writer) {
	usage := "Usage: %s [--exit-after-scan] [--help] [DIRECTORY]\n"
	_, _ = fmt.Fprintf(w, usage, os.Args[0])
}

func printHelp(w io.Writer) {
	printUsage(w)
	desc := "Visually summarize disk usage of DIRECTORY (the current directory by default).\n"
	desc += "\n"
	desc += "Options:\n"
	desc += "      --exit-after-scan  exit after file scanning completes\n"
	desc += "      --help             display this help and exit\n"
	fmt.Fprintln(w, desc)
}

// ArgsOrExit returns valid parameters, or, on either --help or invalid input, exits the program
func ArgsOrExit() (path string, debug bool, exitAfterScan bool) {
	var (
		args       []string = os.Args[1:]
		help       bool
		unknownOpt string
	)

	for _, arg := range args {
		switch {
		case arg == "--help":
			help = true
		case arg == "--debug":
			debug = true
		case arg == "--exit-after-scan":
			exitAfterScan = true
		case strings.HasPrefix(arg, "--"):
			unknownOpt = arg
		default:
			path = arg
		}

		if exit, code := maybeExit(unknownOpt, help); exit {
			os.Exit(code)
		}
	}
	if path == "" {
		path = "."
	}
	return path, debug, exitAfterScan
}

func maybeExit(unknownOpt string, help bool) (exit bool, code int) {
	switch {
	case unknownOpt != "":
		// mimic `git --a` unknown opt behavior:
		// write usage to stderr only on error, otherwise to stdout
		fmt.Fprintln(os.Stderr, "unknown option: "+unknownOpt)
		printUsage(os.Stderr)
		return true, 1
	case help:
		printHelp(os.Stdout)
		return true, 0
	}
	return false, 0
}
