package app

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type EventMouseLocal struct {
	localX, localY int
	event          *tcell.EventMouse
}

func (ev *EventMouseLocal) LocalPosition() (int, int) {
	return ev.localX, ev.localY
}

func (ev *EventMouseLocal) RootEvent() *tcell.EventMouse {
	return ev.event
}

func (ev *EventMouseLocal) When() time.Time {
	return ev.RootEvent().When()
}

func NewEventMouseLocal(ev *tcell.EventMouse, offsetX int, offsetY int) *EventMouseLocal {
	mx, my := ev.Position()
	return &EventMouseLocal{
		localX: mx - offsetX,
		localY: my - offsetY,
		event:  ev,
	}
}
