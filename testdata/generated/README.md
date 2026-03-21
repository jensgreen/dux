# Generated test data

Each subdirectory contains a `files.txt` listing and a `generate-input.sh`
script that captures the listing from its source (e.g. a Docker image).

To generate all fixtures, run from the repo root:

    go generate ./...

This creates a `root/` directory inside each subdirectory containing
null-byte files matching the sizes in `files.txt`. The `root/` directories
are git-ignored.

To add a new fixture, create a new subdirectory with `files.txt` (format:
`{size}\t{path}`) and add a corresponding `go:generate` directive in
`generate.go` at the repo root.
