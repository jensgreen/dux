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

func (ev *EventMouseLocal) NewOffset(offsetX int, offsetY int) *EventMouseLocal {
	return NewEventMouseLocal(
		ev.localX-offsetX,
		ev.localY-offsetY,
		ev.RootEvent(),
	)
}

func (ev *EventMouseLocal) RootEvent() *tcell.EventMouse {
	return ev.event
}

func (ev *EventMouseLocal) When() time.Time {
	return ev.RootEvent().When()
}

func NewEventMouseLocal(localX int, localY int, ev *tcell.EventMouse) *EventMouseLocal {
	return &EventMouseLocal{
		localX: localX,
		localY: localY,
		event:  ev,
	}
}
