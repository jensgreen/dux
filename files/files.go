package files

import (
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

func WalkDir(path string, fileStream chan<- FileEvent, readDir ReadDir) {
	rootInfo, err := os.Stat(path)
	if err != nil {
		fileStream <- FileEvent{Error: err}
	} else if !rootInfo.IsDir() {
		// TODO handle file
		fileStream <- FileEvent{Error: syscall.ENOTDIR}
	} else {
		f := File{
			Path:  path,
			Size:  0,
			IsDir: true,
		}
		log.Println("Sending FileEvent for", f.Path)
		fileStream <- FileEvent{File: f}
		walkDir(path, f, fileStream, readDir)
	}
	log.Println("Closing FileEvent channel")
	close(fileStream)
}

func walkDir(path string, parent File, fileStream chan<- FileEvent, readDir ReadDir) {
	entries, err := readDir(path)
	if err != nil {
		fileStream <- FileEvent{Error: err}
	}

	for _, entry := range entries {
		var size int64 = 0
		path := filepath.Join(parent.Path, entry.Name())
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				fileStream <- FileEvent{Error: err}
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
		fileStream <- FileEvent{File: f}
		if f.IsDir {
			walkDir(f.Path, f, fileStream, readDir)
		}
	}
}
