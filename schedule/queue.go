package schedule

import (
	"errors"
	"sync"
)

var (
	errQueueEmpty    = errors.New("queue empty")
	errLimitExceeded = errors.New("limit exceeded")
)

type queue struct {
	handler func(job)
	limit   int64
	event   chan job
	kill    chan struct{}
	err     chan error
	items   []job
	mutex   *sync.Mutex
	working bool
}

func newQueue(limit int64) *queue {
	q := &queue{
		limit: limit,
		event: make(chan job),
		kill:  make(chan struct{}),
		err:   make(chan error),
		items: make([]job, 0),
		mutex: &sync.Mutex{},
	}
	return q
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

func (q *queue) length() int {
	return len(q.items)
}

func (q *queue) start() {
	for {
		select {
		case e := <-q.event:
			if err := q.push(e); err == errLimitExceeded {
				q.err <- errLimitExceeded
				continue
			}

			if !q.working {
				q.mutex.Lock()
				q.working = true
				go q.handle()
			}

		case <-q.kill:
			break
		}
	}
}

func (q *queue) stop() {
	close(q.kill)
}

func (q *queue) handle() {

	item, err := q.pop()

	if err == errQueueEmpty {
		q.working = false
		q.mutex.Unlock()
		return
	}

	q.handler(item)
	q.handle()
}
