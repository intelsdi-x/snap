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

	"github.com/intelsdi-x/snap/core"
	. "github.com/intelsdi-x/snap/pkg/promise"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

type mockJob struct {
	sync.Mutex

	errors          []error
	worked          bool
	deadline        time.Time
	starttime       time.Time
	completePromise Promise
	numSyncs        int
	rvs             []RendezVous
}

// Create an asynchronous mockJob.
//
// The returned job will NOT block waiting for any calls to RendezVous()
// when Run() is invoked.
func newMockJob() *mockJob {
	return newMultiSyncMockJob(0)
}

// Create a synchronous mockJob.
//
// The returned job WILL block waiting for the supplied number of calls
// to RendezVous() when Run() is invoked before completing.
//
// All callers of Await() will also be transitively blocked by these
// rendez-vous steps.
func newMultiSyncMockJob(n int) *mockJob {
	rvs := make([]RendezVous, n)
	for i := 0; i < n; i++ {
		rvs[i] = NewRendezVous()
	}

	return &mockJob{
		worked:          false,
		deadline:        time.Now().Add(1 * time.Second),
		starttime:       time.Now(),
		completePromise: NewPromise(),
		numSyncs:        0,
		rvs:             rvs,
	}
}

func (mj *mockJob) AddErrors(errs ...error) {
	mj.Lock()
	defer mj.Unlock()
	mj.errors = append(mj.errors, errs...)
}
func (mj *mockJob) Errors() []error      { return mj.errors }
func (mj *mockJob) StartTime() time.Time { return mj.starttime }
func (mj *mockJob) Deadline() time.Time  { return mj.deadline }
func (mj *mockJob) Type() jobType        { return collectJobType }
func (mj *mockJob) TypeString() string   { return "" }
func (mj *mockJob) TaskID() string       { return "" }

// Complete the first incomplete rendez-vous (if there is one)
func (mj *mockJob) RendezVous() {
	mj.Lock()
	defer mj.Unlock()

	if mj.numSyncs < len(mj.rvs) {
		mj.rvs[mj.numSyncs].B()
		mj.numSyncs++
	}
}

func (mj *mockJob) Await() {
	mj.completePromise.Await()
}

func (mj *mockJob) Run() {
	for _, rv := range mj.rvs {
		rv.A()
	}
	mj.worked = true
	mj.completePromise.Complete([]error{})
}

func (mj *mockJob) Name() string {
	return "n/a"
}

func (mj *mockJob) Version() int {
	return 0
}

func (mj *mockJob) Metrics() []core.Metric {
	return nil
}

func TestWorkerManager(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey(".Work()", t, func() {
		Convey("Sends / receives work to / from worker", func() {
			manager := newWorkManager()
			j := newMockJob()
			manager.Work(j)
			j.Await()
			So(j.worked, ShouldEqual, true)
		})

		Convey("does not work job if queuing error occurs", func() {
			log.SetLevel(log.DebugLevel)
			manager := newWorkManager(CollectQSizeOption(1), CollectWkrSizeOption(1))
			manager.Start()

			j1 := newMultiSyncMockJob(2) // j1 will block in Run() twice.
			j2 := newMockJob()
			j3 := newMockJob()

			// Submit three jobs.
			qj1 := manager.Work(j1)
			j1.RendezVous() // First RendezVous with j1 in Run().
			qj2 := manager.Work(j2)
			qj3 := manager.Work(j3)

			// Wait for the third queued job to be marked complete,
			// "out-of-order" and with errors.
			errs3 := qj3.Promise().Await()
			So(errs3, ShouldNotBeEmpty)

			j1.RendezVous() // Second RendezVous with j1 (unblocks j1.Run()).

			errs1 := qj1.Promise().Await()
			So(errs1, ShouldBeEmpty)

			errs2 := qj2.Promise().Await()
			So(errs2, ShouldBeEmpty)

			// The work queue should be empty at this point.
			So(manager.collectq.items, ShouldBeEmpty)

			// The first and second jobs should have been worked.
			So(j1.worked, ShouldBeTrue)
			So(j2.worked, ShouldBeTrue)

			// The third job should have been dropped.
			So(j3.worked, ShouldBeFalse)
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
