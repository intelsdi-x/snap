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
	"errors"

	"github.com/intelsdi-x/snap/pkg/chrono"
	"github.com/pborman/uuid"
)

var workerKillChan = make(chan struct{})

type worker struct {
	id       string
	rcv      <-chan queuedJob
	kamikaze chan struct{}
}

func newWorker(rChan <-chan queuedJob) *worker {
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
		case q := <-w.rcv:
			// assert that deadline is not exceeded
			if chrono.Chrono.Now().Before(q.Job().Deadline()) {
				q.Job().Run()
			} else {
				// the deadline was exceeded and this job will not run
				q.Job().AddErrors(errors.New("Worker refused to run overdue job."))
			}

			// mark the job complete
			q.Promise().Complete(q.Job().Errors())

		// the single kill-channel -- used when resizing worker pools
		case <-w.kamikaze:
			return

		//the broadcast that kills all workers
		case <-workerKillChan:
			return
		}
	}
}
