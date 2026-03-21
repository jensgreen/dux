# dux

A disk usage analyzer. Like [`du(1)`](https://en.wikipedia.org/wiki/Du_(Unix)),
but displays the results as an interactive
[treemap](https://en.wikipedia.org/wiki/Treemapping).

```
testdata 111B (8 files)     (4)
┌testdata/ 111B────────────────────────────────────────────────────────────────┐
│┌example/ 111B───────────────────────────────────────────────────────────────┐│
││┌inner/ 66B─────────────────────────────────┐┌outer.txt 45B────────────────┐││
│││┌a.txt 13B────────────────────────────────┐││                             │││
││││                                         │││                             │││
│││└─────────────────────────────────────────┘││                             │││
│││┌b.txt 38B────────────────────────────────┐││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
││││                                         │││                             │││
│││└─────────────────────────────────────────┘││                             │││
│││┌nested/ 15B──────────────────────────────┐││                             │││
││││┌innermost.txt 15B──────────────────────┐│││                             │││
│││││                                       ││││                             │││
││││└───────────────────────────────────────┘│││                             │││
│││└─────────────────────────────────────────┘││                             │││
││└───────────────────────────────────────────┘└─────────────────────────────┘││
│└────────────────────────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────────────────────┘
```

# Installation

```sh
go install github.com/jensgreen/dux@latest
```

# Usage

```
Usage: dux [--help] [DIRECTORY]
Visually summarize disk usage of DIRECTORY (the current directory by default).

Options:
      --help     display this help and exit
```

Use `+`/`-` to increase/decrease depth, and `q` or `Ctrl-C` to quit.

# Development

```sh
go build ./...   # Build
go test ./...    # Test
```

## Generating test data

Realistic filesystem fixtures can be generated for benchmarking:

```sh
go generate ./...
```

This creates null-byte file trees under `testdata/generated/` matching
real filesystem layouts (e.g. Debian Trixie). See
[`testdata/generated/README.md`](testdata/generated/README.md) for details.
