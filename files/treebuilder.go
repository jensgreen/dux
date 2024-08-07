package files

import (
	"fmt"
	"path/filepath"
)

type TreeBuilder struct {
	root      *FileTree
	pathIndex map[string]*FileTree
}

func (tb *TreeBuilder) Root() (*FileTree, bool) {
	return tb.root, tb.root != nil
}

func (tb *TreeBuilder) Find(path string) (*FileTree, bool) {
	node, ok := tb.pathIndex[path]
	return node, ok
}

// Insert a File to the hierarchy, update weights and relationships
func (tb *TreeBuilder) Insert(f File) error {
	cleanPath := filepath.Clean(f.Path)
	if f.Path != cleanPath {
		return fmt.Errorf("path %q has shorter filepath.Clean equivalent %q", f.Path, cleanPath)
	}

	tree := &FileTree{file: f}
	tb.pathIndex[f.Path] = tree

	if _, ok := tb.Root(); !ok {
		tb.root = tree
		return nil
	}

	parentPath := f.Dir()
	parent, ok := tb.pathIndex[parentPath]
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

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{
		pathIndex: make(map[string]*FileTree),
	}
}
