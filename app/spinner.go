package app

import "time"

var _frames = [...]string{
	"   ",
	".  ",
	".. ",
	"...",
	" ..",
	"  .",
	"   ",
}

type Spinner struct {
	frame      int
	throttle   time.Duration
	lastUpdate time.Time
}

func NewSpinner() *Spinner {
	return &Spinner{
		throttle: 50 * time.Millisecond,
	}
}

func (s *Spinner) String() string {
	return _frames[s.frame]
}

func (s *Spinner) Tick() {
	now := time.Now()
	if now.After(s.lastUpdate.Add(s.throttle)) {
		s.frame = s.nextFrame()
		s.lastUpdate = now
	}
}

func (s *Spinner) nextFrame() int {
	return (s.frame + 1) % len(_frames)
}
