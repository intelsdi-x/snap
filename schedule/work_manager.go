package schedule

import (
	"time"

	"github.com/intelsdilabs/pulse/core"
)

// Job interface
type Job interface {
	Errors() []error
}

type job struct {
	errors []error
}

// NewJob taking []core.MetricType creates and returns a Job
func NewCollectorJob(metricTypes []core.MetricType) Job {
	return &collectorJob{
		metricTypes: metricTypes,
	}
}

// Errors returns the errors that have occured
func (c *job) Errors() []error {
	return c.errors
}

// CollectorJob interface
type CollectorJob interface {
	Job
	Metrics() []core.Metric
}

type collectorJob struct {
	job
	metrics     []core.Metric
	metricTypes []core.MetricType
}

// Metrics returns the metrics
func (c *collectorJob) Metrics() []core.Metric {
	return c.metrics
}

// WorkerManager provides a method to get work done
type ManagesWork interface {
	Work(Job) Job
}

type managesWork struct {
}

// Work dispatch jobs to worker pools for processing
func (w *managesWork) Work(j Job) Job {
	respChan := make(chan Job)
	go func() {
		//TODO send work to worker queue and wait for result
		//results is sent back as a modified job
		// simulate work by just a 500ms sleep
		time.Sleep(time.Millisecond * 500)
		respChan <- j
	}()
	return <-respChan
}
