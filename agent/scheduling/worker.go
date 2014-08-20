package scheduling

import (
	"time"
	"math/rand"
	"github.com/lynxbat/pulse/agent/collection"
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

func (w *MetricWorker) Start(workerChan chan []collection.Metric, quitChan chan bool, ackQuitChan chan bool) {
	for {
		select {
		case <- quitChan:
			// ack the quit signal and exit goroutine
			ackQuitChan <- true
			return
		case metrics := <- workerChan:
			time.Sleep(time.Duration(rand.Int31n(1000)) * time.Millisecond)
			w.MetricsWorked += len(metrics)
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
