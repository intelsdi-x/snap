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

import "sync"

/*

  caller [optional: qj.Promise().Await()]
     ^
     |
     |
  queuedJob
     |
     |
     |                                   job queue
  workManager.Work(job) ---queuedJob---> [jjjjjjj]
                                             |
                                             | workManager.sendToWorker(j)
                                             V
                                        +---------+
                                        | w  w  w |
                  j.Job().Run() +-----  |  w w  w | worker pool
                   j.Complete()         |  w   w  |
                                        +---------+
*/

type workManager struct {
	state          workManagerState
	collectq       *queue
	publishq       *queue
	processq       *queue
	collectWkrs    []*worker
	publishWkrs    []*worker
	processWkrs    []*worker
	collectQSize   uint
	publishQSize   uint
	processQSize   uint
	collectWkrSize uint
	publishWkrSize uint
	processWkrSize uint
	collectchan    chan queuedJob
	publishchan    chan queuedJob
	processchan    chan queuedJob
	kill           chan struct{}
	mutex          *sync.Mutex
}

type workManagerState int

const (
	workManagerStopped workManagerState = iota
	workManagerRunning

	defaultQSize   uint = 5
	defaultWkrSize uint = 1
)

type workManagerOption func(w *workManager) workManagerOption

// CollectQSizeOption sets the collector queue size(length) and
// returns the previous queue option state.
func CollectQSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.collectQSize
		w.collectQSize = v
		return CollectQSizeOption(previous)
	}
}

// PublishQSizeOption sets the publisher queue size(length) and
// returns the previous queue option state.
func PublishQSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.publishQSize
		w.publishQSize = v
		return PublishQSizeOption(previous)
	}
}

// ProcessQSizeOption sets the processor queue size(length) and
// returns the previous queue option state.
func ProcessQSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.processQSize
		w.processQSize = v
		return ProcessQSizeOption(previous)
	}
}

// CollectWkrSizeOption sets the collector worker pool size
// and returns the previous collector worker pool state.
func CollectWkrSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.collectWkrSize
		w.collectWkrSize = v
		return CollectWkrSizeOption(previous)
	}
}

// ProcessWkrSizeOption sets the processor worker pool size
// and return the previous processor worker pool state.
func ProcessWkrSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.processWkrSize
		w.processWkrSize = v
		return ProcessWkrSizeOption(previous)
	}
}

// PublishWkrSizeOption sets the publisher worker pool size
// and returns the previous previous publisher worker pool state.
func PublishWkrSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.publishWkrSize
		w.publishWkrSize = v
		return PublishWkrSizeOption(previous)
	}
}

func newWorkManager(opts ...workManagerOption) *workManager {

	wm := &workManager{
		collectQSize:   defaultQSize,
		processQSize:   defaultQSize,
		publishQSize:   defaultQSize,
		collectWkrSize: defaultWkrSize,
		publishWkrSize: defaultWkrSize,
		processWkrSize: defaultWkrSize,
		collectchan:    make(chan queuedJob),
		publishchan:    make(chan queuedJob),
		processchan:    make(chan queuedJob),
		kill:           make(chan struct{}),
		mutex:          &sync.Mutex{},
	}

	//set options
	for _, opt := range opts {
		opt(wm)
	}

	wm.collectq = newQueue(wm.collectQSize, wm.sendToWorker)
	wm.publishq = newQueue(wm.publishQSize, wm.sendToWorker)
	wm.processq = newQueue(wm.processQSize, wm.sendToWorker)

	wm.publishq.Start()
	wm.collectq.Start()
	wm.processq.Start()

	wm.collectWkrs = make([]*worker, wm.collectWkrSize)
	var i uint
	for i = 0; i < wm.collectWkrSize; i++ {
		wm.collectWkrs[i] = newWorker(wm.collectchan)
		go wm.collectWkrs[i].start()
	}
	wm.publishWkrs = make([]*worker, wm.publishWkrSize)
	for i = 0; i < wm.publishWkrSize; i++ {
		wm.publishWkrs[i] = newWorker(wm.publishchan)
		go wm.publishWkrs[i].start()
	}
	wm.processWkrs = make([]*worker, wm.processWkrSize)
	for i = 0; i < wm.processWkrSize; i++ {
		wm.processWkrs[i] = newWorker(wm.processchan)
		go wm.processWkrs[i].start()
	}
	return wm
}

// Start workManager's loop just handles queuing errors.
func (w *workManager) Start() {

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.state == workManagerStopped {
		w.state = workManagerRunning
		go func() {
			for {
				select {
				case <-w.collectq.Err:
					//TODO: log error
				case <-w.processq.Err:
					//TODO: log error
				case <-w.publishq.Err:
					//TODO: log error
				case <-w.kill:
					return
				}
			}
		}()
	}
}

// Stop closes the collector queue and worker
func (w *workManager) Stop() {
	w.collectq.Stop()
	close(workerKillChan)
	close(w.kill)
}

// Work dispatches jobs to worker pools for processing.
//
// Returns a queued job to the caller, which will be
// completed by the work queue aubsystem.
func (w *workManager) Work(j job) queuedJob {
	qj := newQueuedJob(j)
	switch j.Type() {
	case collectJobType:
		w.collectq.Event <- qj
	case processJobType:
		w.processq.Event <- qj
	case publishJobType:
		w.publishq.Event <- qj
	}
	return qj
}

// AddCollectWorker adds a new worker to
// the collector worker pool
func (w *workManager) AddCollectWorker() {
	nw := newWorker(w.collectchan)
	go nw.start()
	w.collectWkrs = append(w.collectWkrs, nw)
	w.collectWkrSize++
}

// AddPublishWorker adds a new worker to
// the publisher worker pool
func (w *workManager) AddPublishWorker() {
	nw := newWorker(w.publishchan)
	go nw.start()
	w.publishWkrs = append(w.publishWkrs, nw)
	w.publishWkrSize++
}

// AddProcessWorker adds a new worker to
// the processor worker pool
func (w *workManager) AddProcessWorker() {
	nw := newWorker(w.processchan)
	go nw.start()
	w.processWkrs = append(w.processWkrs, nw)
	w.processWkrSize++
}

// sendToWorker is the handler given to the queue.
// it dispatches work to the worker pool.
func (w *workManager) sendToWorker(j queuedJob) {
	switch j.Job().Type() {
	case collectJobType:
		w.collectchan <- j
	case publishJobType:
		w.publishchan <- j
	case processJobType:
		w.processchan <- j
	}
}
