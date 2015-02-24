package schedule

import "github.com/intelsdilabs/pulse/core"

// Job interface
type Job interface {
	Errors() []error
	Metrics() []core.Metric
}

// NewJob taking []core.MetricType creates and returns a Job
func NewJob(metricTypes []core.MetricType) Job {
	return &job{
		metricTypes: metricTypes,
	}
}

type job struct {
	errors      []error
	metrics     []core.Metric
	metricTypes []core.MetricType
}

// Errors returns the errors that have occured
func (c *job) Errors() []error {
	return c.errors
}

// Metrics returns the metrics
func (c *job) Metrics() []core.Metric {
	return c.metrics
}

// WorkerManager provides a method to get work done
type WorkManager interface {
	Work(Job) Job
}

type workManager struct {
}

// Work dispatch jobs to worker pools for processing
func (w *workManager) Work(j Job) Job {
	respChan := make(chan Job)
	go func() {
		//TODO send work to worker queue and wait for result
		//results is sent back as a modified job
		respChan <- j
	}()
	return <-respChan
}

// WorkDispatcher
var WorkDispatcher *workManager = new(workManager)
