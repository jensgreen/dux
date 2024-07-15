package dux

import (
	"fmt"
	"path/filepath"

	"github.com/jensgreen/dux/files"
)

type TreeBuilder struct {
	root       *files.FileTree
	pathLookup map[string]*files.FileTree
	fileCount  map[string]int
}

type FileTreeNode struct {
	Tree      *files.FileTree
	FileCount int
}

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{
		pathLookup: make(map[string]*files.FileTree),
		fileCount:  make(map[string]int),
	}
}

func (tb *TreeBuilder) Root() (FileTreeNode, error) {
	if tb.root == nil {
		return FileTreeNode{}, fmt.Errorf("no root")
	}

	return FileTreeNode{
		Tree:      tb.root,
		FileCount: tb.fileCount[tb.root.File.Path],
	}, nil
}

func (tb *TreeBuilder) FindNode(path string) (FileTreeNode, error) {
	node, ok := tb.pathLookup[path]
	if !ok {
		return FileTreeNode{}, fmt.Errorf("no such node: %s", path)
	}
	return FileTreeNode{
		Tree:      node,
		FileCount: tb.fileCount[path],
	}, nil
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
		tb.fileCount[parent.File.Path] += 1
		path = parentPath
	}
}
