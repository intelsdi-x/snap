package schedule

import (
	"time"

	"code.google.com/p/go-uuid/uuid"
)

var workerKillChan = make(chan struct{})

type worker struct {
	id  string
	rcv <-chan job
}

func newWorker(rChan <-chan job) *worker {
	return &worker{
		rcv: rChan,
		id:  uuid.New(),
	}
}

func (w *worker) start() {
	for {
		select {
		case j := <-w.rcv:
			if time.Now().Unix() < (j.StartTime() + j.Deadline()) {
				j.Run()
			}
		case <-workerKillChan:
			break
		}
	}
}
