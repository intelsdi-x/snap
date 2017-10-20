// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQueue(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("newQueue", t, func() {
		q := newQueue(5, func(queuedJob) {})
		So(q, ShouldHaveSameTypeAs, new(queue))
	})

	Convey("it pops items off and works them", t, func() {
		x := 0
		q := newQueue(5, func(j queuedJob) {
			x = 1
			j.Promise().Complete([]error{})
		})
		q.Start()
		j := &collectorJob{coreJob: &coreJob{}}
		qj := newQueuedJob(j)
		q.Event <- qj
		qj.Promise().Await()
		So(x, ShouldEqual, 1)
		q.Stop()
	})

	Convey("it works the jobs in order", t, func() {
		x := []time.Time{}
		q := newQueue(5, func(j queuedJob) {
			x = append(x, j.Job().Deadline())
			j.Promise().Complete([]error{})
		})
		q.Start()

		numJobs := 4
		var wg sync.WaitGroup
		wg.Add(numJobs)

		for i := 0; i < numJobs; i++ {
			j := &collectorJob{coreJob: &coreJob{}}
			j.deadline = time.Now().Add(time.Duration(i) * time.Second)
			qj := newQueuedJob(j)
			qj.Promise().AndThen(func(errors []error) { wg.Done() })
			q.Event <- qj
		}

		wg.Wait()

		So(x[0].Unix(), ShouldBeLessThan, x[1].Unix())
		So(x[1].Unix(), ShouldBeLessThan, x[2].Unix())
		So(x[2].Unix(), ShouldBeLessThan, x[3].Unix())
		q.Stop()
	})

	Convey("it sends an error if the queue bound is exceeded", t, func() {
		q := newQueue(3, func(queuedJob) { time.Sleep(1 * time.Second) })
		q.Start()
		for i := 0; i < 5; i++ {
			q.Event <- newQueuedJob(&collectorJob{})
		}
		err := <-q.Err
		So(err, ShouldNotBeNil)
		So(err.Err, ShouldResemble, errLimitExceeded)
		q.Stop()
	})

	Convey("stop closes the queue", t, func() {
		q := newQueue(3, func(queuedJob) { time.Sleep(1 * time.Second) })
		q.Start()
		q.Stop()
		time.Sleep(10 * time.Millisecond)
		So(func() { q.kill <- struct{}{} }, ShouldPanic)
		So(func() { q.Event <- newQueuedJob(&collectorJob{}) }, ShouldPanic)
	})

}
