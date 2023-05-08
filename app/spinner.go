package app

import "time"

type Spinner struct {
	currFrame  int
	throttle   time.Duration
	lastUpdate time.Time
	frames     []string
}

func NewSpinner() *Spinner {
	frames := []string{
		"   ",
		".  ",
		".. ",
		"...",
		" ..",
		"  .",
		"   ",
	}

	return &Spinner{
		throttle: 50 * time.Millisecond,
		frames:   frames,
	}
}

func (s *Spinner) String() string {
	return s.frames[s.currFrame]
}

func (s *Spinner) Tick() {
	now := time.Now()
	if now.After(s.lastUpdate.Add(s.throttle)) {
		s.currFrame = s.nextFrame()
		s.lastUpdate = now
	}
}

func (s *Spinner) nextFrame() int {
	return (s.currFrame + 1) % len(s.frames)
}
