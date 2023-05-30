package dux

import (
	"log"
	"path/filepath"

	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/treemap/tiling"
)

type Presenter struct {
	FileEvents  <-chan files.FileEvent
	Commands    <-chan Command
	stateEvents chan<- StateEvent
	Tiler       tiling.Tiler
	state       State
	root        *files.FileTree
	pathLookup  map[string]*files.FileTree
	fileCount   map[string]int
}

func NewPresenter(
	fileEvents <-chan files.FileEvent,
	commands <-chan Command,
	stateEvents chan<- StateEvent,
	initialState State,
	tiler tiling.Tiler,
) Presenter {
	return Presenter{
		FileEvents:  fileEvents,
		Commands:    commands,
		stateEvents: stateEvents,
		state:       initialState,
		Tiler:       tiler,
		pathLookup:  make(map[string]*files.FileTree),
		fileCount:   make(map[string]int),
	}
}

// Add a File to the hierarchy, update weights and relationships
func (p *Presenter) add(f files.File) {
	tree := files.FileTree{File: f}
	parentPath := filepath.Dir(f.Path)
	parent, ok := p.pathLookup[parentPath]
	if ok {
		parent.Children = append(parent.Children, tree)
		p.bubbleUp(f)
		p.pathLookup[f.Path] = &parent.Children[len(parent.Children)-1]
	} else {
		p.root = &tree
		p.pathLookup[f.Path] = &tree
	}
}

func (p *Presenter) bubbleUp(f files.File) {
	var (
		path       string = f.Path
		parentPath string
	)
	for {
		parentPath = filepath.Dir(path)
		parent, ok := p.pathLookup[parentPath]
		// done when there is no parent,
		// or when parent is self (both . and / are their own parents)
		if !ok || path == parentPath {
			return
		}
		// log.Printf("Bubbling up %v to %v", f, parent.File)
		parent.File.Size += f.Size
		p.fileCount[parent.File.Path] += 1
		path = parentPath
	}
}

func (p *Presenter) Loop() {
	for !p.state.Quit {
		p.tick()
	}
}

func (p *Presenter) pollEvent() []error {
	var errs []error
	if p.state.Pause {
		cmd := <-p.Commands
		p.state = p.processCommand(cmd)
	} else {
		select {
		case cmd := <-p.Commands:
			p.state = p.processCommand(cmd)
		case event, ok := <-p.FileEvents:
			if !ok {
				// when closed, never select this channel again
				p.FileEvents = nil
				p.state.IsWalkingFiles = false
				break
			}
			if event.Error != nil {
				errs = append(errs, event.Error)
				break
			}
			f := normalize(event.File)
			log.Printf("Got FileEvent for %v with size %v", f.Path, f.Size)
			p.add(f)
		}
	}
	return errs
}

func (p *Presenter) tick() {
	errs := p.pollEvent()

	if p.root != nil {
		rootRect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, p.state.TreemapSize.AsR2())

		var rootTreemap *treemap.Treemap
		rootFileTree := *p.root
		if p.state.Zoom != nil {
			rootFileTree = *p.pathLookup[p.state.Zoom.Path()]
		}
		rootTreemap = treemap.New(rootFileTree, rootRect, p.Tiler, p.state.MaxDepth, 0)

		if p.state.Selection != nil {
			p.state.Selection = rootTreemap.FindSubTreemap(p.state.Selection.Path())
			p.state.TotalFiles = p.fileCount[p.state.Selection.Path()]
		} else {
			p.state.TotalFiles = p.fileCount[rootFileTree.File.Path]
		}

		if p.state.Zoom != nil {
			p.state.Zoom = rootTreemap.FindSubTreemap(p.state.Zoom.Path())
			p.state.Treemap = p.state.Zoom
		} else {
			p.state.Treemap = rootTreemap
		}
	}

	log.Printf("Sending stateEvent")
	p.stateEvents <- StateEvent{State: p.state, Errors: errs}
	log.Printf("Sent stateEvent")

	if p.state.Quit {
		close(p.stateEvents)
	}
}

func (p *Presenter) processCommand(cmd Command) State {
	log.Printf("Executing command %T", cmd)
	return cmd.Execute(p.state)
}

func normalize(f files.File) files.File {
	f.Path = filepath.Clean(f.Path)
	return f
}
