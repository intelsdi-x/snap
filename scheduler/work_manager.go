/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
                             job queue
    workManager.Work(j) ----> [jjjjjjj]
                   ^                 |
                   |                 | workManager.sendToWorker(j)
                   |                 V
         job.Run() |            +---------+
 <-job.ReplyChan() |            | w  w  w |
                   +----------  |  w w  w | worker pool
                                |  w   w  |
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
	collectchan    chan job
	publishchan    chan job
	processchan    chan job
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

func CollectQSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.collectQSize
		w.collectQSize = v
		return CollectQSizeOption(previous)
	}
}

func PublishQSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.publishQSize
		w.publishQSize = v
		return PublishQSizeOption(previous)
	}
}

func ProcessQSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.processQSize
		w.processQSize = v
		return ProcessQSizeOption(previous)
	}
}

func CollectWkrSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.collectWkrSize
		w.collectWkrSize = v
		return CollectWkrSizeOption(previous)
	}
}

func ProcessWkrSizeOption(v uint) workManagerOption {
	return func(w *workManager) workManagerOption {
		previous := w.processWkrSize
		w.processWkrSize = v
		return ProcessWkrSizeOption(previous)
	}
}

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
		collectchan:    make(chan job),
		publishchan:    make(chan job),
		processchan:    make(chan job),
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

// workManager's loop just handles queuing errors.
func (w *workManager) Start() {

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.state == workManagerStopped {
		w.state = workManagerRunning
		go func() {
			for {
				select {
				case qe := <-w.collectq.Err:
					qe.Job.ReplChan() <- struct{}{}
					//TODO: log error
				case qe := <-w.processq.Err:
					qe.Job.ReplChan() <- struct{}{}
					//TODO: log error
				case qe := <-w.publishq.Err:
					qe.Job.ReplChan() <- struct{}{}
					//TODO: log error
				case <-w.kill:
					return
				}
			}
		}()
	}
}

func (w *workManager) Stop() {
	w.collectq.Stop()
	close(workerKillChan)
	close(w.kill)
}

// Work dispatches jobs to worker pools for processing.
// a job is queued, a worker receives it, and then replies
// on the job's  reply channel.
func (w *workManager) Work(j job) job {
	switch j.Type() {
	case collectJobType:
		w.collectq.Event <- j
	case processJobType:
		w.processq.Event <- j
	case publishJobType:
		w.publishq.Event <- j
	}
	<-j.ReplChan()
	return j
}

func (w *workManager) AddCollectWorker() {
	nw := newWorker(w.collectchan)
	go nw.start()
	w.collectWkrs = append(w.collectWkrs, nw)
	w.collectWkrSize++
}

func (w *workManager) AddPublishWorker() {
	nw := newWorker(w.publishchan)
	go nw.start()
	w.publishWkrs = append(w.publishWkrs, nw)
	w.publishWkrSize++
}

func (w *workManager) AddProcessWorker() {
	nw := newWorker(w.processchan)
	go nw.start()
	w.processWkrs = append(w.processWkrs, nw)
	w.processWkrSize++
}

// sendToWorker is the handler given to the queue.
// it dispatches work to the worker pool.
func (w *workManager) sendToWorker(j job) {
	switch j.Type() {
	case collectJobType:
		w.collectchan <- j
	case publishJobType:
		w.publishchan <- j
	case processJobType:
		w.processchan <- j
	}
}
