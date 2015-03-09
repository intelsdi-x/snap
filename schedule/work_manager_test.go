package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type mockJob struct {
	worked    bool
	replchan  chan struct{}
	deadline  int64
	starttime int64
}

func (mj *mockJob) Errors() []error         { return []error{} }
func (mj *mockJob) StartTime() int64        { return mj.starttime }
func (mj *mockJob) Deadline() int64         { return mj.deadline }
func (mj *mockJob) Type() jobType           { return collectJobType }
func (mj *mockJob) ReplChan() chan struct{} { return mj.replchan }

func (mj *mockJob) Run() {
	mj.worked = true
	mj.replchan <- struct{}{}
}

func TestWorkerManager(t *testing.T) {
	Convey(".Work()", t, func() {
		Convey("Sends / receives work to / from worker", func() {
			manager = newWorkManager(int64(5), 1)
			var j job
			j = &mockJob{
				worked:    false,
				replchan:  make(chan struct{}),
				deadline:  int64(1),
				starttime: time.Now().Unix(),
			}
			j = manager.Work(j)
			So(j.(*mockJob).worked, ShouldEqual, true)
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

}
