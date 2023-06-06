package files

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

type ReadDir = func(dirname string) ([]os.DirEntry, error)

type File struct {
	Path  string
	Size  int64
	IsDir bool
}

func (f *File) Name() string {
	return filepath.Base(f.Path)
}

type FileTree struct {
	File     File
	Children []FileTree
}

type FileEvent struct {
	File  File
	Error error
}

func WalkDir(ctx context.Context, path string, fileEvents chan<- FileEvent, readDir ReadDir) {
	defer close(fileEvents)
	rootInfo, err := os.Stat(path)
	if err != nil {
		send(ctx, fileEvents, FileEvent{Error: err})
	} else if !rootInfo.IsDir() {
		// TODO handle file
		send(ctx, fileEvents, FileEvent{Error: syscall.ENOTDIR})
	} else {
		f := File{
			Path:  path,
			Size:  0,
			IsDir: true,
		}
		log.Println("Sending FileEvent for", f.Path)
		send(ctx, fileEvents, FileEvent{File: f})
		walkDir(ctx, path, f, fileEvents, readDir)
	}
	log.Println("Closing FileEvent channel")
}

func walkDir(ctx context.Context, path string, parent File, fileEvents chan<- FileEvent, readDir ReadDir) {
	entries, err := readDir(path)
	if err != nil {
		send(ctx, fileEvents, FileEvent{Error: err})
	}

	for _, entry := range entries {
		var size int64 = 0
		path := filepath.Join(parent.Path, entry.Name())
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				send(ctx, fileEvents, FileEvent{Error: err})
				continue
			} else {
				size = info.Size()
			}
		}
		f := File{
			Path:  path,
			Size:  size,
			IsDir: entry.IsDir(),
		}
		log.Println("Sending FileEvent for", f.Path)
		send(ctx, fileEvents, FileEvent{File: f})
		if f.IsDir {
			walkDir(ctx, f.Path, f, fileEvents, readDir)
		}
	}
}

func send(ctx context.Context, ch chan<- FileEvent, ev FileEvent) {
	select {
	case <-ctx.Done():
	case ch <- ev:
	}
}
