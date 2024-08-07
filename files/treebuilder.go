package files

import (
	"fmt"
	"path/filepath"
)

type TreeBuilder struct {
	root       *FileTree
	pathLookup map[string]*FileTree
}

func (tb *TreeBuilder) Root() (*FileTree, error) {
	if tb.root == nil {
		return nil, fmt.Errorf("no root")
	}
	return tb.root, nil
}

func (tb *TreeBuilder) FindNode(path string) (*FileTree, error) {
	node, ok := tb.pathLookup[path]
	if !ok {
		return nil, fmt.Errorf("no such node: %s", path)
	}
	return node, nil
}

// Insert a File to the hierarchy, update weights and relationships
func (tb *TreeBuilder) Insert(f File) error {
	cleanPath := filepath.Clean(f.Path)
	if f.Path != cleanPath {
		return fmt.Errorf("path %q has shorter filepath.Clean equivalent %q", f.Path, cleanPath)
	}

	tree := &FileTree{file: f}
	tb.pathLookup[f.Path] = tree

	if _, err := tb.Root(); err != nil {
		tb.root = tree
		return nil
	}

	parentPath := f.Dir()
	parent, ok := tb.pathLookup[parentPath]
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
		pathLookup: make(map[string]*FileTree),
	}
}
