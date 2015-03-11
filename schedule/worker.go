package schedule

import (
	"time"

	"code.google.com/p/go-uuid/uuid"
)

var workerKillChan = make(chan struct{})

type worker struct {
	id       string
	rcv      <-chan job
	kamikaze chan struct{}
}

func newWorker(rChan <-chan job) *worker {
	return &worker{
		rcv:      rChan,
		id:       uuid.New(),
		kamikaze: make(chan struct{}),
	}
}

// begin a worker
func (w *worker) start() {
	for {
		select {
		case j := <-w.rcv:
			// assert that deadline is not exceeded
			if time.Since(j.StartTime()) < j.Deadline() {
				j.Run()
				continue
			}
			// reply immediately -- Job not run
			j.ReplChan() <- struct{}{}

		// the single kill-channel -- used when resizing worker pools
		case <-w.kamikaze:
			return

		//the broadcast that kills all workers
		case <-workerKillChan:
			return

		}
	}
}
