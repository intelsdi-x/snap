package schedule

import "fmt"

type workManager struct {
	collectq       *queue
	collectqSize   int64
	collectWkrs    []*worker
	collectWkrSize int
	collectchan    chan job
}

func newWorkManager(cqs int64, cws int) *workManager {

	wm := &workManager{
		collectq:       newQueue(cqs),
		collectWkrSize: cws,
		collectchan:    make(chan job),
	}

	wm.collectq.Handler = wm.sendToWorker
	go wm.collectq.Start()

	wm.collectWkrs = make([]*worker, cws)
	for i := 0; i < cws; i++ {
		wm.collectWkrs[i] = newWorker(wm.collectchan)
		go wm.collectWkrs[i].start()
	}

	return wm
}

func (w *workManager) Start() {
	for {
		select {
		case <-w.collectq.Err:
			// TODO(dpitt): handle queuing error
		case <-workerKillChan:
			fmt.Println("STOPPING WM")
			return
		}
	}
}

func (w *workManager) Stop() {
	w.collectq.Stop()
	fmt.Println("CLOSING KILL CHAN")
	close(workerKillChan)
}

// Work dispatches jobs to worker pools for processing
func (w *workManager) Work(j job) job {
	switch j.Type() {
	case collectJobType:
		w.collectq.Event <- j
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

func (w *workManager) sendToWorker(j job) {
	switch j.Type() {
	case collectJobType:
		w.collectchan <- j
	}
}
