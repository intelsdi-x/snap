package schedule

import (
	"errors"
	"sync"
)

const (
	queueStopped queueStatus = iota
	queueRunning
	queueWorking
)

var (
	errQueueEmpty    = errors.New("queue empty")
	errLimitExceeded = errors.New("limit exceeded")
)

type queuingError struct {
	Job job
	Err error
}

func (qe *queuingError) Error() string {
	return qe.Err.Error()
}

type queue struct {
	Event   chan job
	Handler func(job)
	Err     chan *queuingError

	limit  int64
	kill   chan struct{}
	items  []job
	mutex  *sync.Mutex
	status queueStatus
}

type queueStatus int

func newQueue(limit int64) *queue {
	q := &queue{
		Event: make(chan job),
		Err:   make(chan *queuingError),

		limit:  limit,
		kill:   make(chan struct{}),
		items:  make([]job, 0),
		mutex:  &sync.Mutex{},
		status: queueStopped,
	}
	return q
}

func (q *queue) Start() {

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.status == queueStopped {
		q.status = queueRunning
		go q.start()
	}
}

func (q *queue) Stop() {
	q.mutex.Lock()
	if q.status != queueStopped {
		close(q.kill)
		q.status = queueStopped
	}
	q.mutex.Unlock()
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

			if q.status == queueRunning {
				q.status = queueWorking
				go q.handle()
			}

		case <-q.kill:
			// this "officially" closes the Event channel.
			// after this, an attempt to write to a stopped queue will panic.
			// otherwise, a goroutine will sleep forever.
			go func() { close(q.Event) }()
			<-q.Event
			return
		}
	}

}

func (q *queue) handle() {

	item, err := q.pop()

	if err == errQueueEmpty {
		q.status = queueRunning
		return
	}

	q.Handler(item)
	q.handle()
}

func (q *queue) length() int {
	return len(q.items)
}

func (q *queue) push(j job) error {
	if q.limit == 0 || int64(q.length()+1) <= q.limit {
		q.items = append(q.items, j)
		return nil
	}
	return errLimitExceeded
}

func (q *queue) pop() (job, error) {

	var j job

	if q.items == nil || len(q.items) == 0 {
		return j, errQueueEmpty
	}

	j = q.items[0]
	q.items = q.items[1:]

	return j, nil
}
