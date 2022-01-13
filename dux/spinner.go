package dux

import "time"

// TODO(jensgreen): package structure could be improved so that the sccope of
// Spinner and _frames (among others) is smaller
var _frames = [...]string{
	"   ",
	".  ",
	".. ",
	"...",
	" ..",
	"  .",
	"   ",
}

type spinner struct {
	frame      int
	throttle   time.Duration
	lastUpdate time.Time
}

func newSpinner() *spinner {
	return &spinner{
		throttle: 50 * time.Millisecond,
	}
}

func (s *spinner) String() string {
	return _frames[s.frame]
}

func (s *spinner) Tick() {
	now := time.Now()
	if now.After(s.lastUpdate.Add(s.throttle)) {
		s.frame = s.nextFrame()
		s.lastUpdate = now
	}
}

func (s *spinner) nextFrame() int {
	var i = s.frame + 1
	if i == len(_frames) {
		return 0
	}
	return i
}
