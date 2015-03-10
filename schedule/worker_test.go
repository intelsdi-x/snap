package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorker(t *testing.T) {
	Convey("Replies on the Job's reply chan", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan job)
		w := newWorker(rcv)
		go w.start()
		mj := &mockJob{
			replchan:  make(chan struct{}),
			starttime: time.Now().Unix(),
			deadline:  int64(1000),
		}
		rcv <- mj
		So(mj.worked, ShouldEqual, true)
	})
	Convey("replies without running job if deadline is exceeded", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan job)
		w := newWorker(rcv)
		go w.start()
		mj := &mockJob{
			replchan:  make(chan struct{}),
			starttime: time.Now().Unix(),
			deadline:  int64(1),
		}
		time.Sleep(time.Millisecond * 1500)
		rcv <- mj
		<-mj.replchan
		So(mj.worked, ShouldEqual, false)
	})
	Convey("stops the worker if kamikaze chan is closed", t, func() {
		workerKillChan = make(chan struct{})
		rcv := make(chan job)
		w := newWorker(rcv)
		go func() { close(w.kamikaze) }()
		w.start()
		So(0, ShouldEqual, 0)
	})
}
