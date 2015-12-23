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
	"sync"
)

var (
	errQueueEmpty    = errors.New("queue empty")
	errLimitExceeded = errors.New("limit exceeded")
)

type jobHandler func(queuedJob)

type queue struct {
	Event chan queuedJob
	Err   chan *queuingError

	handler jobHandler
	limit   uint
	kill    chan struct{}
	items   []queuedJob
	mutex   *sync.Mutex
	status  queueStatus
}

type queueStatus int

const (
	queueStopped queueStatus = iota // queue not running
	queueRunning                    // queue running, but not working. the queue must be in the is state before entering handle
	queueWorking                    // queue is currently being worked (a goroutine is currently inside q.handle())
)

type queuingError struct {
	Job job
	Err error
}

func (qe *queuingError) Error() string {
	return qe.Err.Error()
}

func newQueue(limit uint, handler jobHandler) *queue {
	return &queue{
		Event: make(chan queuedJob),
		Err:   make(chan *queuingError),

		handler: handler,
		limit:   limit,
		kill:    make(chan struct{}),
		items:   []queuedJob{},
		mutex:   &sync.Mutex{},
		status:  queueStopped,
	}
}

// begins the queue handling loop
func (q *queue) Start() {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.status == queueStopped {
		q.status = queueRunning
		go q.start()
	}
}

// Stop closes both Err and Event channels, and
// causes the handling loop to exit.
func (q *queue) Stop() {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.status != queueStopped {
		close(q.kill)
		q.status = queueStopped
	}
}

/*
   Below is the private, internal functionality of the queue.
   These functions are not thread-safe, and should not be used
   outside the queue itself.  The only interaction between a queue
   and outside consumers should be through the Event chan, the
   Err chan, Start(), or Stop().
*/

func (q *queue) start() {
	for {
		select {
		case e := <-q.Event:
			if err := q.push(e); err != nil {
				qe := &queuingError{
					Err: err,
					Job: e.Job(),
				}
				q.Err <- qe
				e.Promise().Complete([]error{qe}) // Signal job termination.
				continue
			}

			q.mutex.Lock()
			if q.status == queueRunning {
				q.status = queueWorking
				q.mutex.Unlock()
				go q.handle()
				continue
			}
			q.mutex.Unlock()

		case <-q.kill:
			// this "officially" closes the Event channel.
			// after this, an attempt to write to a stopped queue will panic.
			// otherwise, a goroutine will sleep forever, waiting for a reader
			// of Event.
			go func() { close(q.Event) }()
			<-q.Event
			return
		}
	}

}

func (q *queue) handle() {
	for {
		item, err := q.pop()
		if err == errQueueEmpty {
			q.mutex.Lock()
			q.status = queueRunning
			q.mutex.Unlock()
			return
		}
		q.handler(item)
	}
}

func (q *queue) length() int {
	return len(q.items)
}

func (q *queue) push(j queuedJob) error {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.limit == 0 || uint(q.length())+1 <= q.limit {
		q.items = append(q.items, j)
		return nil
	}
	return errLimitExceeded
}

func (q *queue) pop() (queuedJob, error) {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	var j queuedJob

	if q.length() == 0 {
		return j, errQueueEmpty
	}

	j = q.items[0]
	q.items = q.items[1:]

	return j, nil
}
