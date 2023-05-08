package dux

import (
	"log"
	"path/filepath"

	"github.com/jensgreen/dux/files"
)

type Presenter struct {
	FileEvents  <-chan files.FileEvent
	Commands    <-chan Command
	stateEvents chan<- StateEvent
	Tiler       Tiler
	state       State
	root        *files.FileTree
	pathLookup  map[string]*files.FileTree
}

func NewPresenter(
	fileEvents <-chan files.FileEvent,
	commands <-chan Command,
	stateEvents chan<- StateEvent,
	initialState State,
	tiler Tiler,
) Presenter {
	return Presenter{
		FileEvents:  fileEvents,
		Commands:    commands,
		stateEvents: stateEvents,
		state:       initialState,
		Tiler:       tiler,
		pathLookup:  make(map[string]*files.FileTree),
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
		path = parentPath
	}
}

func (p *Presenter) Loop() {
	for !p.state.Quit {
		p.tick()
	}
}

func (p *Presenter) tick() {
	var errs []error
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
		p.state.TotalFiles = len(p.pathLookup)
	}
	if p.root != nil {
		rootTreemap := TreemapWithTiler(*p.root, p.state.TreemapRect, p.Tiler, p.state.MaxDepth, 0)
		p.state.Treemap = rootTreemap
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
