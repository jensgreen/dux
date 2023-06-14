package dux

import (
	"context"
	"log"
	"path/filepath"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/treemap/tiling"
)

type Presenter struct {
	ctx         context.Context
	shutdown         context.CancelFunc
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
	ctx context.Context,
	shutdown context.CancelFunc,
	fileEvents <-chan files.FileEvent,
	commands <-chan Command,
	stateEvents chan<- StateEvent,
	initialState State,
	tiler tiling.Tiler,
) Presenter {
	return Presenter{
		ctx:         ctx,
		shutdown:    shutdown,
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
		select {
		case <-p.ctx.Done():
			p.state.Quit = true
			return nil
		case cmd := <-p.Commands:
			p.state = p.processCommand(cmd)
		}
	} else {
		select {
		case <-p.ctx.Done():
			p.state.Quit = true
			return nil
		case cmd := <-p.Commands:
			log.Printf("Presenter got command %T", cmd)
			p.state = p.processCommand(cmd)
		case event, ok := <-p.FileEvents:
			log.Printf("Presenter got FileEvent")
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
	defer func() {
		if p.state.Quit {
			log.Printf("Presenter: Quit: enter")
			close(p.stateEvents)
			log.Printf("Presenter: Quit: calling shutdown func")
			p.shutdown()
			log.Printf("Presenter: Quit: done")
		}
	}()

	errs := p.pollEvent()
	if p.root != nil {
		rootRect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, z2.PointAsR2(p.state.TreemapSize))

		var rootTreemap *treemap.R2Treemap
		rootFileTree := *p.root
		if p.state.Zoom != nil {
			rootFileTree = *p.pathLookup[p.state.Zoom.Path()]
		}
		rootTreemap = treemap.NewR2Treemap(rootFileTree, rootRect, p.Tiler, p.state.MaxDepth)

		if p.state.Selection != nil {
			selection, err := rootTreemap.FindNode(p.state.Selection.Path())
			if err != nil {
				// selected node has been removed from the new treemap
				// TODO select closest (grand)parent still remaining
				p.state.Selection = nil
				p.state.TotalFiles = p.fileCount[rootFileTree.File.Path]
			} else {
				p.state.Selection = selection
				p.state.TotalFiles = p.fileCount[p.state.Selection.Path()]
			}
		} else {
			p.state.TotalFiles = p.fileCount[rootFileTree.File.Path]
		}

		if p.state.Zoom != nil {
			zoom, err := rootTreemap.FindNode(p.state.Zoom.Path())
			if err != nil {
				// selected zoom node has been removed from the new treemap
				// TODO select closest (grand)parent still remaining
				p.state.Zoom = nil
				p.state.Treemap = rootTreemap
			} else {
				p.state.Zoom = zoom
				p.state.Treemap = p.state.Zoom
			}
		} else {
			p.state.Treemap = rootTreemap
		}
	}

	log.Printf("Sending stateEvent")
	select {
	case <-p.ctx.Done():
		p.state.Quit = true
		return
	case p.stateEvents <- StateEvent{State: p.state, Errors: errs}:
		log.Printf("Sent stateEvent")
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
