package scheduler

import (
	"errors"
	"sync"
)

var (
	errQueueEmpty    = errors.New("queue empty")
	errLimitExceeded = errors.New("limit exceeded")
)

type queue struct {
	Event chan job
	Err   chan *queuingError

	handler func(job)
	limit   int64
	kill    chan struct{}
	items   []job
	mutex   *sync.Mutex
	status  queueStatus
}

type queueStatus int

const (
	queueStopped queueStatus = iota // queue not running
	queueRunning                    // queue running, but not working. the queue must be in the is state before entering handle
	queueWorking                    // queue is currently being worked (a goroutine is currently inside q.handle())
)

type jobHandler func(job)

type queuingError struct {
	Job job
	Err error
}

func (qe *queuingError) Error() string {
	return qe.Err.Error()
}

func newQueue(limit int64, handler jobHandler) *queue {
	return &queue{
		Event: make(chan job),
		Err:   make(chan *queuingError),

		handler: handler,
		limit:   limit,
		kill:    make(chan struct{}),
		items:   []job{},
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
					Job: e,
				}
				q.Err <- qe
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

func (q *queue) push(j job) error {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.limit == 0 || int64(q.length()+1) <= q.limit {
		q.items = append(q.items, j)
		return nil
	}
	return errLimitExceeded
}

func (q *queue) pop() (job, error) {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	var j job

	if q.length() == 0 {
		return j, errQueueEmpty
	}

	j = q.items[0]
	q.items = q.items[1:]

	return j, nil
}
