package files

import (
	"fmt"
	"path/filepath"
)

type FS struct {
	root       *FileTree
	pathLookup map[string]*FileTree
}

func (fs *FS) Root() (*FileTree, bool) {
	return fs.root, fs.root != nil
}

func (fs *FS) Find(path string) (*FileTree, bool) {
	node, ok := fs.pathLookup[path]
	return node, ok
}

// Insert a File to the hierarchy, update weights and relationships
func (fs *FS) Insert(f File) error {
	cleanPath := filepath.Clean(f.Path)
	if f.Path != cleanPath {
		return fmt.Errorf("path %q has shorter filepath.Clean equivalent %q", f.Path, cleanPath)
	}

	tree := &FileTree{file: f}
	fs.pathLookup[f.Path] = tree

	if _, ok := fs.Root(); !ok {
		fs.root = tree
		return nil
	}

	parentPath := f.Dir()
	parent, ok := fs.pathLookup[parentPath]
	if ok {
		tree.SetParent(parent)
		parent.AddChildren(tree)
	}

	for ; ok; parent, ok = parent.Parent() {
		parent.file.Size += f.Size
		parent.file.NumDescendants++
	}
	return nil
}

func NewFS() *FS {
	return &FS{
		pathLookup: make(map[string]*FileTree),
	}
}
