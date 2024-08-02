package files

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/jensgreen/dux/cancellable"
)

type ReadDir = func(dirname string) ([]os.DirEntry, error)

type File struct {
	Path           string
	Size           int64
	IsDir          bool
	NumDescendants int
}

func (f *File) Name() string {
	return filepath.Base(f.Path)
}

type FileTree struct {
	file     File
	parent   *FileTree
	children []*FileTree
}

func (ft *FileTree) File() File {
	return ft.file
}

func (ft *FileTree) Parent() (*FileTree, bool) {
	return ft.parent, ft.parent != nil
}

func (ft *FileTree) Children() []*FileTree {
	return ft.children[:]
}

func (ft *FileTree) SetParent(parent *FileTree) {
	ft.parent = parent
}

func (ft *FileTree) AddChildren(children ...*FileTree) {
	ft.children = append(ft.children, children...)
}

func NewFileTree(f File) *FileTree {
	return &FileTree{file: f}
}

type FileEvent struct {
	File  File
	Error error
}

func WalkDir(ctx context.Context, path string, fileEvents chan<- FileEvent, readDir ReadDir) {
	defer func() {
		log.Println("Closing FileEvent channel")
		close(fileEvents)
	}()

	rootInfo, err := os.Stat(path)
	if err != nil {
		err := cancellable.Send(ctx, fileEvents, FileEvent{Error: err})
		if err != nil {
			return
		}
	} else if !rootInfo.IsDir() {
		// TODO handle file
		err := cancellable.Send(ctx, fileEvents, FileEvent{Error: syscall.ENOTDIR})
		if err != nil {
			return
		}
	} else {
		f := File{
			Path:  path,
			Size:  0,
			IsDir: true,
		}
		log.Println("Sending FileEvent for", f.Path)
		err := cancellable.Send(ctx, fileEvents, FileEvent{File: f})
		if err != nil {
			return
		}

		err = walkDir(ctx, path, f, fileEvents, readDir)
		if err != nil {
			return
		}
	}
}

func walkDir(ctx context.Context, path string, parent File, fileEvents chan<- FileEvent, readDir ReadDir) error {
	entries, err := readDir(path)
	if err != nil {
		err := cancellable.Send(ctx, fileEvents, FileEvent{Error: err})
		if err != nil {
			return err
		}
	}

	for _, entry := range entries {
		var size int64 = 0
		path := filepath.Join(parent.Path, entry.Name())
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				err := cancellable.Send(ctx, fileEvents, FileEvent{Error: err})
				if err != nil {
					return err
				}
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
		err := cancellable.Send(ctx, fileEvents, FileEvent{File: f})
		if err != nil {
			return err
		}
		if f.IsDir {
			err := walkDir(ctx, f.Path, f, fileEvents, readDir)
			if errors.Is(cancellable.ErrClosed, err) {
				return err
			}
		}
	}
	return nil
}
