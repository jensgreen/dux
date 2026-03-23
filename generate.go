// This file exists solely to hold go:generate directives. The actual generator
// script is in generate_testdata.go, which uses //go:build ignore to stay out
// of the normal build. Since ignored files are also invisible to go generate,
// the directive must live here in a file the toolchain processes.
package main

//go:generate bash -c "cat testdata/generated/debian-trixie/files.txt | go run generate_testdata.go testdata/generated/debian-trixie/root"
