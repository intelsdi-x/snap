package schedule

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
			manager := newWorkManager(int64(5), 1)
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
			manager := newWorkManager(int64(1), 1)
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
				errors:    []error{errors.New("j3")},
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
	})
	Convey("Stop()", t, func() {
		Convey("Stops the queue and the workers", func() {
			mgr := newWorkManager(int64(5), 1)
			go mgr.Start()
			mgr.Stop()
			So(mgr.collectq.status, ShouldEqual, queueStopped)
		})
	})
	Convey("AddCollectWorker()", t, func() {
		Convey("it adds a collect worker", func() {
			mgr := newWorkManager(int64(5), 1)
			mgr.AddCollectWorker()
			So(mgr.collectWkrSize, ShouldEqual, 2)
			So(mgr.collectWkrSize, ShouldEqual, len(mgr.collectWkrs))
		})
	})
}
