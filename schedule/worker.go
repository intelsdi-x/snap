package schedule

import (
	"fmt"
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

func (w *worker) start() {
	for {
		select {
		case j := <-w.rcv:
			if time.Now().Unix() < (j.StartTime() + j.Deadline()) {
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
			fmt.Println("KILLING WORKER")
			return

		}
	}
}
