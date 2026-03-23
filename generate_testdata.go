//go:build ignore

// Generates test fixture directory trees from file listings. File listings
// are stored in testdata/generated/ — see generate-input.sh for how they are
// captured.
//
// Run via: go generate ./...
package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const usage = `Usage: go run generate_testdata.go DIRECTORY

Reads file tree structure from stdin and writes matching null-byte file tree to DIRECTORY.
If DIRECTORY already exists, verifies that its contents match the input instead.

Input line format: {SIZE}\t{path}\n
`

type fileEntry struct {
	Size uint64
	Path string
}

func parseInput(scanner *bufio.Scanner, outdir string) []fileEntry {
	var entries []fileEntry
	i := 0
	for scanner.Scan() {
		i++
		line := scanner.Text()
		before, after, found := strings.Cut(line, "\t")

		if !found {
			fmt.Fprintf(os.Stderr, "line %d: bad format: %s\n", i, line)
			os.Exit(1)
		}

		size, err := strconv.ParseUint(before, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "line %d: bad size: %s: %s\n", i, before, err.Error())
			os.Exit(1)
		}

		path := filepath.Join(outdir, filepath.Clean(after))
		entries = append(entries, fileEntry{Size: size, Path: path})
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return entries
}

func generate(entries []fileEntry) {
	// Create dirs
	for _, e := range entries {
		dirname, _ := filepath.Split(e.Path)
		if err := os.MkdirAll(dirname, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "could not create dir: %s\n", err.Error())
			os.Exit(1)
		}
	}
	// Write files
	for _, e := range entries {
		nullbuf := make([]byte, e.Size)
		if err := os.WriteFile(e.Path, nullbuf, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "could not write file: %s\n", err.Error())
			os.Exit(1)
		}
	}
}

func verify(entries []fileEntry, outdir string) {
	ok := true
	expected := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		expected[e.Path] = struct{}{}
		info, err := os.Stat(e.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "missing: %s\n", e.Path)
			ok = false
			continue
		}
		if uint64(info.Size()) != e.Size {
			fmt.Fprintf(os.Stderr, "size mismatch: %s: want %d, got %d\n", e.Path, e.Size, info.Size())
			ok = false
		}
	}
	filepath.Walk(outdir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if _, found := expected[p]; !found {
			fmt.Fprintf(os.Stderr, "unexpected: %s\n", p)
			ok = false
		}
		return nil
	})
	if !ok {
		fmt.Fprintf(os.Stderr, "verification failed; delete %s and re-run to regenerate\n", outdir)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	outdir := path.Clean(os.Args[1])
	scanner := bufio.NewScanner(os.Stdin)
	entries := parseInput(scanner, outdir)

	if _, err := os.Stat(outdir); err == nil {
		verify(entries, outdir)
	} else {
		generate(entries)
	}
}
