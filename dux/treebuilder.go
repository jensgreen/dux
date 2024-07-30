package dux

import (
	"fmt"
	"path/filepath"

	"github.com/jensgreen/dux/files"
)

type TreeBuilder struct {
	root       *files.FileTree
	pathLookup map[string]*files.FileTree
}

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{
		pathLookup: make(map[string]*files.FileTree),
	}
}

func (tb *TreeBuilder) Root() (*files.FileTree, error) {
	if tb.root == nil {
		return nil, fmt.Errorf("no root")
	}
	return tb.root, nil
}

func (tb *TreeBuilder) FindNode(path string) (*files.FileTree, error) {
	node, ok := tb.pathLookup[path]
	if !ok {
		return nil, fmt.Errorf("no such node: %s", path)
	}
	return node, nil
}

// Add a File to the hierarchy, update weights and relationships
func (tb *TreeBuilder) Add(f files.File) {
	tree := files.FileTree{File: f}
	parentPath := filepath.Dir(f.Path)
	parent, ok := tb.pathLookup[parentPath]
	if ok {
		parent.Children = append(parent.Children, tree)
		tb.bubbleUp(f)
		tb.pathLookup[f.Path] = &parent.Children[len(parent.Children)-1]
	} else {
		tb.root = &tree
		tb.pathLookup[f.Path] = &tree
	}
}

func (tb *TreeBuilder) bubbleUp(f files.File) {
	var (
		path       string = f.Path
		parentPath string
	)
	for {
		parentPath = filepath.Dir(path)
		parent, ok := tb.pathLookup[parentPath]
		// done when there is no parent,
		// or when parent is self (both . and / are their own parents)
		if !ok || path == parentPath {
			return
		}
		// log.Printf("Bubbling up %v to %v", f, parent.File)
		parent.File.Size += f.Size
		parent.File.NumDescendants++
		path = parentPath
	}
}
