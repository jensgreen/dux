package recovery

// Recovery can be used to defer panics, allowing other goroutines a chance
// to clean up.
type Recovery interface {
	// Go launches given func in a goroutine, and defers panics until Release()
	// is called.
	Go(func())
	// Release should be called when cleanup is done to re-panic a recovered panic.
	Release()
}

type recovery struct {
	cancel  func()
	release chan struct{}
}

func (r *recovery) Go(f func()) {
	go func() {
		defer r.recover()
		f()
	}()
}

func (r *recovery) Release() {
	close(r.release)
}

func (r *recovery) recover() {
	if p := recover(); p != nil {
		r.cancel()
		<-r.release
		panic(p)
	}
}

func New(cancel func()) *recovery {
	return &recovery{
		cancel:  cancel,
		release: make(chan struct{}),
	}
}
