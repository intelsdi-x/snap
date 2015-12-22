package promise

import (
	"sync"
	"testing"

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
}
