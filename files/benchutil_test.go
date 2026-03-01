package files

import (
	"fmt"
	"path/filepath"
)

// buildBalancedTree creates a synthetic FileTree with the given branching
// factor and depth. Each leaf gets a size of 1; interior nodes accumulate
// their descendants' sizes. This is useful for benchmarking treemap
// construction, tiling, and navigation with realistic tree shapes.
func buildBalancedTree(basePath string, branching, depth int) *FileTree {
	size := leafCount(branching, depth)
	root := NewFileTree(File{Path: basePath, Size: int64(size), IsDir: true, NumDescendants: size - 1})
	if depth > 0 {
		addChildren(root, basePath, branching, depth-1)
	}
	return root
}

func addChildren(parent *FileTree, parentPath string, branching, remainingDepth int) {
	for i := 0; i < branching; i++ {
		childPath := filepath.Join(parentPath, fmt.Sprintf("d%d", i))
		if remainingDepth == 0 {
			child := NewFileTree(File{Path: childPath, Size: 1, IsDir: false})
			child.SetParent(parent)
			parent.AddChildren(child)
		} else {
			size := leafCount(branching, remainingDepth)
			child := NewFileTree(File{Path: childPath, Size: int64(size), IsDir: true, NumDescendants: size - 1})
			child.SetParent(parent)
			parent.AddChildren(child)
			addChildren(child, childPath, branching, remainingDepth-1)
		}
	}
}

// leafCount returns branching^depth (the number of leaves in a full tree).
func leafCount(branching, depth int) int {
	n := 1
	for i := 0; i < depth; i++ {
		n *= branching
	}
	return n
}

// buildWideTree creates a tree with a single level of n children.
func buildWideTree(basePath string, n int) *FileTree {
	root := NewFileTree(File{Path: basePath, Size: int64(n), IsDir: true, NumDescendants: n})
	for i := 0; i < n; i++ {
		child := NewFileTree(File{Path: filepath.Join(basePath, fmt.Sprintf("f%d", i)), Size: 1})
		child.SetParent(root)
		root.AddChildren(child)
	}
	return root
}

// buildDeepTree creates a linear chain of depth nodes.
func buildDeepTree(basePath string, depth int) *FileTree {
	root := NewFileTree(File{Path: basePath, Size: 1, IsDir: true, NumDescendants: depth})
	current := root
	path := basePath
	for i := 0; i < depth; i++ {
		path = filepath.Join(path, fmt.Sprintf("d%d", i))
		child := NewFileTree(File{Path: path, Size: 1, IsDir: true, NumDescendants: depth - i - 1})
		child.SetParent(current)
		current.AddChildren(child)
		current = child
	}
	return current // return the deepest node so callers can walk up
}

// populateFS inserts n files into an FS using a flat directory structure.
func populateFS(basePath string, n int) *FS {
	fs := NewFS()
	fs.Insert(File{Path: basePath, Size: 0, IsDir: true})
	for i := 0; i < n; i++ {
		fs.Insert(File{
			Path: filepath.Join(basePath, fmt.Sprintf("f%d", i)),
			Size: int64(i + 1),
		})
	}
	return fs
}

// populateFSDeep inserts files forming a chain of the given depth.
func populateFSDeep(basePath string, depth int) *FS {
	fs := NewFS()
	path := basePath
	fs.Insert(File{Path: path, Size: 0, IsDir: true})
	for i := 0; i < depth; i++ {
		path = filepath.Join(path, fmt.Sprintf("d%d", i))
		fs.Insert(File{Path: path, Size: 1, IsDir: true})
	}
	return fs
}
