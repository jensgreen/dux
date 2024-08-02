package files

type FileNode interface {
	File() File
	Parent() (*FileNode, bool)
	Children() []*FileNode

	// setParent(*File)
	// addChildren(...*File)
}

type FSTree interface {
	Root() (FileNode, bool)
	Find(path string) (FileNode, bool)
	Insert(File)
}
