package promise

import (
	"sync"
)

// A disposable write-once latch, to act as a synchronization
// barrier to signal completion of some asynchronous operation
// (successful or otherwise).
//
// Functions that operate on this type (IsComplete, Complete,
// Await) are idempotent and thread-safe.
type Promise interface {
	IsComplete() bool
	Complete(errors []error)
	Await() []error
	AndThen(f func([]error))
}

func NewPromise() Promise {
	return &promise{
		complete:     false,
		completeChan: make(chan struct{}),
	}
}

type promise struct {
	sync.Mutex

	complete     bool
	errors       []error
	completeChan chan struct{}
}

// Returns whether this promise is complete yet, without blocking.
func (p *promise) IsComplete() bool {
	return p.complete
}

// Unblock all goroutines awaiting promise completion.
func (p *promise) Complete(errors []error) {
	p.Lock()
	defer p.Unlock()

	if !p.complete {
		p.complete = true
		p.errors = errors
		close(p.completeChan)
	}
}

// Blocks the caller until the promise is marked complete.
func (p *promise) Await() []error {
	<-p.completeChan
	return p.errors
}

// Invokes the supplied function after this promise completes.
func (p *promise) AndThen(f func([]error)) {
	go func() {
		f(p.Await())
	}()
}
