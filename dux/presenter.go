package dux

import (
	"context"
	"log"

	"github.com/jensgreen/dux/cancellable"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/treemap/tiling"
)

type Presenter struct {
	ctx         context.Context
	shutdown    context.CancelFunc
	FileEvents  <-chan files.FileEvent
	Commands    <-chan Command
	stateEvents chan<- StateEvent
	Tiler       tiling.Tiler
	state       State
	fs          *files.FS
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
		fs:          files.NewFS(), // TODO inject dep
	}
}

func (p *Presenter) Loop() {
	for !p.state.Quit {
		p.tick()
	}
}

func (p *Presenter) pollEvent() (Action, []error) {
	var errs []error
	var action Action
	if p.state.Pause {
		cmd, err := cancellable.Receive(p.ctx, p.Commands)
		if err != nil {
			p.state.Quit = true
			return ActionNone, nil
		}
		p.state, action = p.processCommand(cmd)
	} else {
		select {
		case <-p.ctx.Done():
			p.state.Quit = true
			return ActionNone, nil
		case cmd := <-p.Commands:
			log.Printf("Presenter got command %T", cmd)
			p.state, action = p.processCommand(cmd)
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
			f := event.File
			log.Printf("Got FileEvent for %v with size %v", f.Path, f.Size)
			p.fs.Insert(f)
		}
	}
	return action, errs
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

	action, errs := p.pollEvent()
	root, ok := p.fs.Root()
	if ok {
		rootRect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, z2.PointAsR2(p.state.TreemapSize))

		var rootTreemap *treemap.R2Treemap
		var rootFileTree = *root
		if p.state.Zoom != nil {
			node, ok := p.fs.Find(p.state.Zoom.Path())
			if !ok {
				panic("zoom target removed from tree")
			}
			rootFileTree = *node
		}
		rootTreemap = treemap.NewR2Treemap(rootFileTree, rootRect, p.Tiler, p.state.MaxDepth)

		if p.state.Selection != nil {
			selection, err := rootTreemap.FindNode(p.state.Selection.Path())
			if err != nil {
				// selected node has been removed from the new treemap
				// TODO select closest (grand)parent still remaining
				p.state.Selection = nil
				p.state.TotalFiles = 1 + root.File().NumDescendants
			} else {
				node, ok := p.fs.Find(p.state.Selection.Path())
				if !ok {
					panic("selection removed from tree")
				}
				p.state.Selection = selection
				p.state.TotalFiles = 1 + node.File().NumDescendants
			}
		} else {
			p.state.TotalFiles = 1 + root.File().NumDescendants
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
	err := cancellable.Send(p.ctx, p.stateEvents, StateEvent{State: p.state, Action: action, Errors: errs})
	if err != nil {
		p.state.Quit = true
		return
	}
	log.Printf("Sent stateEvent")
}

func (p *Presenter) processCommand(cmd Command) (State, Action) {
	log.Printf("Executing command %T", cmd)
	return cmd.Execute(p.state)
}
