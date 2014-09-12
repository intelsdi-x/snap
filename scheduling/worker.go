package scheduling

import (
	"time"
	"math/rand"
	"code.google.com/p/go-uuid/uuid"
)

// Works scheduled metric tasks assigned
type MetricWorker struct {
	// This is lazy loaded to avoid having to use a constructor
	// so only access via the UUID() reader method
	uuid *string
	WorkReceived int
	MetricsWorked int
}

type MetricWorkerJob struct {

}

func (w *MetricWorker) Start(workerChan chan work, quitChan chan bool, ackQuitChan chan bool) {
	for {
		select {
		case <- quitChan:
			// ack the quit indicating this worker will no longer process tasks
			ackQuitChan <- true
			// Return and exit function
			return
		case work := <- workerChan:

			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			w.MetricsWorked += len(work.Metrics)
			w.WorkReceived++
		}
	}
}

// Returns unique ID
// This lazy loads the ID on the first load
func (w *MetricWorker) UUID() string{
	if w.uuid == nil {
		u := uuid.New()
		w.uuid = &u
	}
	return *w.uuid
}
