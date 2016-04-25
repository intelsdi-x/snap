// +build legacy

package promise

import (
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPromise(t *testing.T) {
	Convey("IsComplete()", t, func() {
		Convey("it should return the completion status", func() {
			p := NewPromise()
			So(p.IsComplete(), ShouldBeFalse)
			p.Complete([]error{})
			So(p.IsComplete(), ShouldBeTrue)
		})
	})
	Convey("IsError()", t, func() {
		Convey("it should return the error status", func() {
			Convey("after completing without errors, IsError() returns false", func() {
				p := NewPromise()
				So(p.IsError(), ShouldBeFalse)
				p.Complete([]error{})
				So(p.IsError(), ShouldBeFalse)
			})
			Convey("after completing with errors, IsError() returns true", func() {
				p := NewPromise()
				So(p.IsError(), ShouldBeFalse)
				p.Complete([]error{errors.New("ERROR")})
				So(p.IsError(), ShouldBeTrue)
			})
		})
	})
	Convey("Complete()", t, func() {
		Convey("it should unblock any waiting goroutines", func() {
			p := NewPromise()

			numWaiters := 3
			var wg sync.WaitGroup
			wg.Add(numWaiters)

			for i := 0; i < numWaiters; i++ {
				go func() {
					Convey("all waiting goroutines should see empty errors", t, func() {
						errors := p.Await()
						So(errors, ShouldBeEmpty)
						wg.Done()
					})
				}()
			}

			p.Complete([]error{})
			wg.Wait()
		})
	})
	Convey("AwaitUntil()", t, func() {
		Convey("it should return with errors on timeout", func() {
			p := NewPromise()
			errors := p.AwaitUntil(time.Nanosecond)
			So(errors, ShouldNotBeEmpty)
		})
	})
	Convey("AndThen()", t, func() {
		Convey("it should defer the supplied closure until after completion", func() {
			p := NewPromise()

			funcRan := false
			c := make(chan struct{})

			p.AndThen(func(errors []error) {
				funcRan = true
				close(c)
			})

			// The callback should not have been executed yet.
			So(funcRan, ShouldBeFalse)

			// Trigger callback execution by completing the queued job.
			p.Complete([]error{})

			// Wait for the deferred function to be executed.
			<-c
			So(funcRan, ShouldBeTrue)
		})
	})
	Convey("AndThenUntil()", t, func() {
		Convey("it should defer the supplied closure until timeout", func() {
			p := NewPromise()
			timeout := time.Nanosecond

			var resultErrors []error
			c := make(chan struct{})

			callback := func(errors []error) {
				resultErrors = errors
				close(c)
			}

			p.AndThenUntil(timeout, callback)

			// Wait for the deferred function to be executed.
			<-c
			So(resultErrors, ShouldNotBeEmpty)
			expectedMsg := "Await timed out for promise after [1ns]"
			So(resultErrors[0].Error(), ShouldEqual, expectedMsg)
		})
	})
}

func TestRendezVous(t *testing.T) {
	Convey("IsComplete()", t, func() {
		Convey("it should return the completion status", func() {
			r := NewRendezVous()
			So(r.IsComplete(), ShouldBeFalse)
			go r.A()
			r.B()
			So(r.IsComplete(), ShouldBeTrue)
		})
	})
	Convey("A() and B()", t, func() {
		Convey("it should synchronize concurrent processes", func() {
			r1, r2 := NewRendezVous(), NewRendezVous()
			evidence := false

			go func() {
				r1.A()
				evidence = true
				r2.A()
			}()

			r1.B()
			r2.B()
			So(evidence, ShouldBeTrue)
		})
	})
}
