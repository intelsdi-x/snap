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

type mockJob struct {
	errors         []error
	worked         bool
	replchan       chan struct{}
	rendezvouschan chan struct{}
	deadline       time.Time
	starttime      time.Time
}

func (mj *mockJob) Errors() []error         { return mj.errors }
func (mj *mockJob) StartTime() time.Time    { return mj.starttime }
func (mj *mockJob) Deadline() time.Time     { return mj.deadline }
func (mj *mockJob) Type() jobType           { return collectJobType }
func (mj *mockJob) ReplChan() chan struct{} { return mj.replchan }
func (mj *mockJob) RendezVousStart() {
	<-mj.rendezvouschan
}
func (mj *mockJob) RendezVousEnd() {
	mj.rendezvouschan <- struct{}{}
}

func (mj *mockJob) Run() {
	// Signal for rendez-vous.
	mj.rendezvouschan <- struct{}{}

	// Busy-wait for rendez-vous response.
	for {
		select {
		case <-mj.rendezvouschan:
			mj.worked = true
			mj.replchan <- struct{}{}
			return
		default:
			continue
		}
	}
}

func TestWorkerManager(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey(".Work()", t, func() {
		Convey("Sends / receives work to / from worker", func() {
			manager := newWorkManager()
			j := &mockJob{
				worked:         false,
				replchan:       make(chan struct{}),
				rendezvouschan: make(chan struct{}),
				deadline:       time.Now().Add(1 * time.Second),
				starttime:      time.Now(),
			}
			go manager.Work(j)
			j.RendezVousStart()
			j.RendezVousEnd()
			So(j.worked, ShouldEqual, true)
		})

		Convey("does not work job if queuing error occurs", func() {
			log.SetLevel(log.DebugLevel)
			manager := newWorkManager(CollectQSizeOption(1), CollectWkrSizeOption(1))
			manager.Start()

			var jobs [3]mockJob
			for i := range jobs {
				jobs[i] = mockJob{
					errors:         []error{},
					worked:         false,
					replchan:       make(chan struct{}),
					rendezvouschan: make(chan struct{}),
					deadline:       time.Now().Add(1 * time.Second),
					starttime:      time.Now(),
				}
			}
			j1 := &jobs[0]
			j2 := &jobs[1]
			j3 := &jobs[2]

			completechan := make(chan struct{})

			releaseAsync := func(j *mockJob) {
				manager.Work(j)
				for {
					select {
					case completechan <- struct{}{}: // Signal completion.
						return
					default:
						continue
					}
				}
			}

			// Release one async job and wait for it to be scheduled.
			go releaseAsync(j1)
			j1.RendezVousStart()

			// Release two more async jobs.
			go releaseAsync(j2)
			go releaseAsync(j3)

			// We don't know which of j2 or j3 will hit the queue first.
			// However, because all three jobs share the same rendez-vous channel
			// as soon as one is scheduled this goroutine will unblock.
			go func() {
				// Wait for one of j2 or j3 to be scheduled.
				select {
				case <-j2.rendezvouschan:
				case <-j3.rendezvouschan:
				}

				// Allow the one that got scheduled to continue.
				select {
				case j2.rendezvouschan <- struct{}{}:
				case j3.rendezvouschan <- struct{}{}:
				}
			}()

			// Let j1 continue.
			j1.RendezVousEnd()

			// Await two receives on `completechan`:
			// - one for j1 to complete.
			// - one for the second job, j2 or j3, to complete or drop.
			<-completechan
			<-completechan

			// Rationale: we know j1 was enqueued before j2 and j3 were released.
			So(j1.worked, ShouldEqual, true)

			// Rationale: at least one of j2 or j3 was dropped.
			So(len(manager.collectq.items), ShouldEqual, 0)

			// Rationale: zero or one of j2 or j3 was worked.
			So((!j2.worked || !j3.worked), ShouldEqual, true)
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
