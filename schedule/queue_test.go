package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQueue(t *testing.T) {
	Convey("newQueue", t, func() {
		q := newQueue(5, func(job) {})
		So(q, ShouldHaveSameTypeAs, new(queue))
	})

	Convey("it receives jobs and adds them to queue", t, func() {
		q := newQueue(5, func(job) { time.Sleep(1 * time.Second) })
		q.Start()
		q.Event <- &collectorJob{}
		So(q.length(), ShouldEqual, 1)
		q.Stop()
	})

	Convey("it pops items off and works them", t, func() {
		x := 0
		q := newQueue(5, func(job) { x = 1 })
		q.Start()
		q.Event <- &collectorJob{}
		time.Sleep(time.Millisecond * 10)
		So(x, ShouldEqual, 1)
		q.Stop()
	})

	Convey("it works the jobs in order", t, func() {
		x := []time.Duration{}
		q := newQueue(5, func(j job) { x = append(x, j.Deadline()) })
		q.Start()
		for i := 0; i < 4; i++ {
			j := &collectorJob{}
			j.deadline = time.Duration(time.Duration(i) * time.Second)
			q.Event <- j
		}
		time.Sleep(time.Millisecond * 10)
		So(x, ShouldResemble, []time.Duration{time.Duration(0), time.Duration(1 * time.Second), time.Duration(2 * time.Second), time.Duration(3 * time.Second)})
		q.Stop()
	})

	Convey("it sends an error if the queue bound is exceeded", t, func() {
		q := newQueue(3, func(job) { time.Sleep(1 * time.Second) })
		q.Start()
		for i := 0; i < 5; i++ {
			q.Event <- &collectorJob{}
		}
		err := <-q.Err
		So(err, ShouldNotBeNil)
		So(err.Err, ShouldResemble, errLimitExceeded)
		q.Stop()
	})

	Convey("stop closes the queue", t, func() {
		q := newQueue(3, func(job) { time.Sleep(1 * time.Second) })
		q.Start()
		q.Stop()
		time.Sleep(10 * time.Millisecond)
		So(func() { q.kill <- struct{}{} }, ShouldPanic)
		So(func() { q.Event <- &collectorJob{} }, ShouldPanic)
	})

}
