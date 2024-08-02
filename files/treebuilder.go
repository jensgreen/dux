package files

import (
	"fmt"
	"path/filepath"
)

type TreeBuilder struct {
	root       *FileTree
	pathLookup map[string]*FileTree
}

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{
		pathLookup: make(map[string]*FileTree),
	}
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

// Add a File to the hierarchy, update weights and relationships
func (tb *TreeBuilder) Add(f File) {
	tree := &FileTree{file: f}
	parentPath := filepath.Dir(f.Path)
	parent, ok := tb.pathLookup[parentPath]
	if ok {
		parent.children = append(parent.children, tree)
		tb.bubbleUp(f)
		tb.pathLookup[f.Path] = parent.children[len(parent.children)-1]
	} else {
		tb.root = tree
		tb.pathLookup[f.Path] = tree
	}
}

func (tb *TreeBuilder) bubbleUp(f File) {
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
		parent.file.Size += f.Size
		parent.file.NumDescendants++
		path = parentPath
	}
}

func Normalize(f File) File {
	f.Path = filepath.Clean(f.Path)
	return f
}
