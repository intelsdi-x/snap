package schedule

type workManager struct {
	collectq        *queue
	collectqSize    int64
	collectWkrs     []*worker
	collectWrkrSize int
	collectchan     chan job
}

func newWorkManager(cqs int64, cws int) *workManager {

	wm := &workManager{
		collectq:        newQueue(cqs),
		collectWrkrSize: cws,
		collectchan:     make(chan job),
	}

	wm.collectq.handler = wm.sendToWorker
	go wm.collectq.start()

	wm.collectWkrs = make([]*worker, cws)
	for i := 0; i < cws; i++ {
		wm.collectWkrs[i] = newWorker(wm.collectchan)
		go wm.collectWkrs[i].start()
	}

	return wm
}

func (w *workManager) start() {
	for {
		select {
		case <-w.collectq.err:
			// TODO(dpitt): handle queuing error
		}
	}
}

func (w *workManager) stop() {
	w.collectq.stop()
	close(workerKillChan)
}

// Work dispatches jobs to worker pools for processing
func (w *workManager) Work(j job) job {
	switch j.Type() {
	case collectJobType:
		w.collectq.event <- j
	}
	<-j.ReplChan()
	return j
}

func (w *workManager) sendToWorker(j job) {
	switch j.Type() {
	case collectJobType:
		w.collectchan <- j
	}
}

func (w *workManager) addCollectWorker() {
	nw := newWorker(w.collectchan)
	go nw.start()
}
