package schedule

type workManager struct {
	collectq       *queue
	collectqSize   int64
	collectWkrs    []*worker
	collectWkrSize int
	collectchan    chan job
	kill           chan struct{}
}

func newWorkManager(cqs int64, cws int) *workManager {

	wm := &workManager{
		collectq:       newQueue(cqs),
		collectWkrSize: cws,
		collectchan:    make(chan job),
		kill:           make(chan struct{}),
	}

	wm.collectq.Handler = wm.sendToWorker
	wm.collectq.Start()

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
		case <-w.kill:
			return
		}
	}
}

func (w *workManager) Stop() {
	w.collectq.Stop()
	close(workerKillChan)
	close(w.kill)
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
