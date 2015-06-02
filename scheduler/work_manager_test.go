package scheduler

import (
	"errors"

	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type mockJob struct {
	errors    []error
	worked    bool
	replchan  chan struct{}
	deadline  time.Time
	starttime time.Time
}

func (mj *mockJob) Errors() []error         { return mj.errors }
func (mj *mockJob) StartTime() time.Time    { return mj.starttime }
func (mj *mockJob) Deadline() time.Time     { return mj.deadline }
func (mj *mockJob) Type() jobType           { return collectJobType }
func (mj *mockJob) ReplChan() chan struct{} { return mj.replchan }

func (mj *mockJob) Run() {

	mj.worked = true
	time.Sleep(time.Millisecond * 100)
	mj.replchan <- struct{}{}
}

func TestWorkerManager(t *testing.T) {
	Convey(".Work()", t, func() {
		Convey("Sends / receives work to / from worker", func() {
			manager := newWorkManager()
			var j job
			j = &mockJob{
				worked:    false,
				replchan:  make(chan struct{}),
				deadline:  time.Now().Add(1 * time.Second),
				starttime: time.Now(),
			}
			j = manager.Work(j)

			So(j.(*mockJob).worked, ShouldEqual, true)
		})
		Convey("does not work job if queuing error occurs", func() {
			manager := newWorkManager()
			manager.Start()
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
				errors:    []error{},
				worked:    false,
				replchan:  make(chan struct{}),
				deadline:  time.Now().Add(1 * time.Second),
				starttime: time.Now(),
			}
			go manager.Work(j1)

			go manager.Work(j2)
			manager.Work(j3)

			time.Sleep(time.Millisecond * 10)
			worked := 0
			for _, j := range []*mockJob{j1, j2, j3} {

				if j.worked == true {
					worked++
				}
			}
			So(worked, ShouldEqual, 2)
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
