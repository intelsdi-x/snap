package promise

import (
	"fmt"
	"sync"
	"time"
)

// A disposable write-once latch, to act as a synchronization
// barrier to signal completion of some asynchronous operation
// (successful or otherwise).
//
// Functions that operate on this type (IsComplete, Complete,
// Await, AwaitUntil) are idempotent and thread-safe.
type Promise interface {
	IsComplete() bool
	Complete(errors []error)
	Await() []error
	AwaitUntil(timeout time.Duration) []error
	AndThen(f func([]error))
	AndThenUntil(timeout time.Duration, f func([]error))
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

// Blocks the caller until the promise is marked complete. This function
// is equivalent to invoking AwaitUntil() with a zero-length duration.
// To avoid blocking the caller indefinitely, use AwaitUntil() with a
// non-zero duration instead.
func (p *promise) Await() []error {
	return p.AwaitUntil(0 * time.Second)
}

// Blocks the caller until the promise is marked complete, or the supplied
// duration has elapsed. If the promise has not been completed before the
// await times out, this function returns with nonempty errors. If the
// supplied duration has zero length, this await will never time out.
func (p *promise) AwaitUntil(duration time.Duration) []error {
	var timeoutChan <-chan time.Time
	if duration.Nanoseconds() > 0 {
		timeoutChan = time.After(duration)
	}

	select {
	case <-p.completeChan:
		return p.errors
	case <-timeoutChan:
		return []error{fmt.Errorf("Await timed out for promise after [%s]", duration)}
	}
}

// Invokes the supplied function after this promise completes. This function
// is equivalent to invoking AndThenUntil() with a zero-length duration.
// To avoid blocking a goroutine indefinitely, use AndThenUntil() with a
// non-zero duration instead.
func (p *promise) AndThen(f func([]error)) {
	p.AndThenUntil(0*time.Nanosecond, f)
}

// Invokes the supplied function after this promise completes or times out
// after the supplied duration. If the supplied duration has zero length,
// the deferred execution will never time out.
func (p *promise) AndThenUntil(d time.Duration, f func([]error)) {
	go func() {
		f(p.AwaitUntil(d))
	}()
}

// A reciprocal promise that makes it easy for two coordinating routines
// A and B to wait on each other before proceeding.
type RendezVous interface {
	IsComplete() bool
	A()
	B()
}

func NewRendezVous() RendezVous {
	return &rendezVous{
		a: NewPromise(),
		b: NewPromise(),
	}
}

type rendezVous struct {
	a Promise
	b Promise
}

// Returns whether this rendez-vous is complete yet, without blocking.
func (r *rendezVous) IsComplete() bool {
	return r.a.IsComplete() && r.b.IsComplete()
}

// Complete process A's half of the rendez-vous, and block until process
// B has done the same.
func (r *rendezVous) A() {
	r.a.Complete([]error{})
	r.b.Await()
}

// Complete process B's half of the rendez-vous, and block until process
// A has done the same.
func (r *rendezVous) B() {
	r.b.Complete([]error{})
	r.a.Await()
}
