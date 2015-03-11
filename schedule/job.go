package schedule

import (
	"time"

	"github.com/intelsdilabs/pulse/core"
)

const (
	collectJobType jobType = iota
)

const (
	defaultDeadline = time.Duration(5 * time.Second)
)

// Primary type for job inside
// the scheduler.  Job encompasses all
// all job types -- collect, process, and publish.
type job interface {
	Errors() []error
	StartTime() time.Time
	Deadline() time.Duration
	Type() jobType
	ReplChan() chan struct{}
	Run()
}

type jobType int

// CollectorJob interface
type collectJob interface {
	job
	Metrics() []core.Metric
}

type collectorJob struct {
	jtype       jobType
	deadline    time.Duration
	starttime   time.Time
	errors      []error
	metrics     []core.Metric
	metricTypes []core.MetricType
	replchan    chan struct{}
}

func newCollectorJob(metricTypes []core.MetricType) *collectorJob {
	return &collectorJob{
		jtype:       collectJobType,
		deadline:    defaultDeadline,
		metricTypes: metricTypes,
		metrics:     []core.Metric{},
		errors:      make([]error, 0),
		starttime:   time.Now(),
		replchan:    make(chan struct{}),
	}
}

func (c *collectorJob) StartTime() time.Time {
	return c.starttime
}

func (c *collectorJob) Deadline() time.Duration {
	return c.deadline
}

func (c *collectorJob) Type() jobType {
	return c.jtype
}

func (c *collectorJob) ReplChan() chan struct{} {
	return c.replchan
}

func (c *collectorJob) Metrics() []core.Metric {
	return c.metrics
}

func (c *collectorJob) Errors() []error {
	return c.errors
}

func (c *collectorJob) Run() {
	//ret := metricManager.Collect(c.metrics)
	//c.values = ret
	c.replchan <- struct{}{}
}
