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
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const usage = `Usage: go run gentestdata.go DIRECTORY

Reads file tree structure from stdin and writes matching null-byte file tree to DIRECTORY.

Input line format: {SIZE}\t{path}\n
`

type fileSize struct {
	Size uint64
	Path string
}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err.Error())
		}
	}()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	outdir := os.Args[1]
	empty, err := isEmpty(outdir)
	if !empty || err != nil {
		fmt.Fprintf(os.Stderr, "output directory must exist and be empty\n")
		os.Exit(1)
	}
	outdir = path.Clean(outdir)

	scanner := bufio.NewScanner(os.Stdin)
	i := 0

	// Parse input
	parsed := make([]fileSize, 0)
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
		parsed = append(parsed, fileSize{Size: size, Path: path})
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Create dirs
	for _, fs := range parsed {
		dirname, _ := filepath.Split(fs.Path)
		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create dir: %s\n", err.Error())
			os.Exit(1)
		}
	}

	// Write files
	for _, fs := range parsed {
		nullbuf := make([]byte, fs.Size)
		err := os.WriteFile(fs.Path, nullbuf, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not write file: %s\n", err.Error())
			os.Exit(1)
		}
	}
}
