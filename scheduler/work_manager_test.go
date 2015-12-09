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
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

// A mockJob can be either synchronus (will not return from Run()
// until a receiver is ready on completeChan) or asynchronous.
//
// Test code can rendez-vous with synchronous mockJobs in Run()
// by invoking Await(), which also unblocks the worker executing
// the job.
//
// For asynchronous mockJobs, Await() blocks the caller until the
// job does a buffered send on completeChan in Run().
type mockJob struct {
	errors       []error
	worked       bool
	deadline     time.Time
	starttime    time.Time
	completeChan chan struct{}
}

func newMockJob(sync bool) *mockJob {
	var completeChan chan struct{}
	if sync {
		completeChan = make(chan struct{})
	} else {
		completeChan = make(chan struct{}, 1)
	}

	return &mockJob{
		worked:       false,
		deadline:     time.Now().Add(1 * time.Second),
		starttime:    time.Now(),
		completeChan: completeChan,
	}
}

func (mj *mockJob) Errors() []error      { return mj.errors }
func (mj *mockJob) StartTime() time.Time { return mj.starttime }
func (mj *mockJob) Deadline() time.Time  { return mj.deadline }
func (mj *mockJob) Type() jobType        { return collectJobType }
func (mj *mockJob) Await()               { <-mj.completeChan }

func (mj *mockJob) Run() {
	mj.worked = true
	mj.completeChan <- struct{}{}
}

func TestWorkerManager(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey(".Work()", t, func() {
		Convey("Sends / receives work to / from worker", func() {
			manager := newWorkManager()
			j := newMockJob(false)
			manager.Work(j)
			j.Await()
			So(j.worked, ShouldEqual, true)
		})

		Convey("does not work job if queuing error occurs", func() {
			log.SetLevel(log.DebugLevel)
			manager := newWorkManager(CollectQSizeOption(1), CollectWkrSizeOption(1))
			manager.Start()

			j1 := newMockJob(true) // j1 does a blocking send on its completeChan
			j2 := newMockJob(false)
			j3 := newMockJob(false)

			// Submit three jobs.
			qjs := []queuedJob{}
			qjs = append(qjs, manager.Work(j1))
			qjs = append(qjs, manager.Work(j2))
			qjs = append(qjs, manager.Work(j3))

			// Await completion of j1 (also unblocking j1.Run()).
			j1.Await()

			// Wait for all queued jobs to be marked complete.
			for _, qj := range qjs {
				qj.Await()
			}

			// The work queue should be empty at this point.
			So(manager.collectq.items, ShouldBeEmpty)

			// The first job should have been worked.
			So(j1.worked, ShouldBeTrue)

			// At least one of the second and third jobs should have been dropped.
			So(j2.worked && j3.worked, ShouldBeFalse)
		})

		// The below convey is WIP
		/*Convey("Collect queue error ", func() {
			wMOption1 := CollectQSizeOption(1)
			manager := newWorkManager(wMOption1)
			manager.Start()
			So(manager.collectQSize, ShouldResemble, uint(1))
			j1 := &mockJob{
				errors:    []error{errors.New("j1")},
				worked:    false,
				replchan:  make(chan struct{}),
				deadline:  time.Now().Add(1 * time.Second),
				starttime: time.Now(),
			}
			j2 := &mockJob{
				errors:    []error{errors.New("j2")},
				worked:    false,
				replchan:  make(chan struct{}),
				deadline:  time.Now().Add(1 * time.Second),
				starttime: time.Now(),
			}
			j3 := &mockJob{
				errors:    []error{errors.New("j3")},
				worked:    false,
				replchan:  make(chan struct{}),
				deadline:  time.Now().Add(1 * time.Second),
				starttime: time.Now(),
			}
			manager.Work(j1)

			go manager.Work(j2)
			manager.Work(j3)
			go manager.Work(j1)

			manager.Work(j2)
			go manager.Work(j3)
			manager.Work(j1)

			manager.Work(j2)
			manager.Work(j3)
		})*/
		Convey("testing workMangerOptions", func() {
			wMOption1 := CollectQSizeOption(100)
			wMOption2 := PublishQSizeOption(100)
			wMOption3 := CollectWkrSizeOption(100)
			wMOption4 := ProcessQSizeOption(100)
			wMOption5 := ProcessWkrSizeOption(100)
			wMOption6 := PublishWkrSizeOption(100)
			manager := newWorkManager(wMOption1, wMOption2, wMOption3, wMOption4, wMOption5, wMOption6)
			manager.Start()
			manager.AddPublishWorker()
			manager.AddProcessWorker()
			manager.AddCollectWorker()

			So(manager.collectQSize, ShouldResemble, uint(100))
			So(manager.publishQSize, ShouldResemble, uint(100))
			So(manager.collectWkrSize, ShouldResemble, uint(101))
			So(manager.processWkrSize, ShouldResemble, uint(101))
			So(manager.publishWkrSize, ShouldResemble, uint(101))
			So(manager.publishQSize, ShouldResemble, uint(100))

		})
	})

	Convey("Stop()", t, func() {
		Convey("Stops the queue and the workers", func() {
			mgr := newWorkManager()
			go mgr.Start()
			mgr.Stop()
			So(mgr.collectq.status, ShouldEqual, queueStopped)
		})
	})
	Convey("AddCollectWorker()", t, func() {
		Convey("it adds a collect worker", func() {
			mgr := newWorkManager()
			mgr.AddCollectWorker()
			So(mgr.collectWkrSize, ShouldEqual, 2)
			So(mgr.collectWkrSize, ShouldEqual, len(mgr.collectWkrs))
		})
	})
}
