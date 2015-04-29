package scheduler

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/ctypes"
	"github.com/intelsdilabs/pulse/pkg/logger"
)

const (
	collectJobType jobType = iota
	publishJobType
	processJobType
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
	Deadline() time.Time
	Type() jobType
	ReplChan() chan struct{}
	Run()
}

type jobType int

type coreJob struct {
	jtype     jobType
	deadline  time.Time
	starttime time.Time
	errors    []error
	replchan  chan struct{}
}

func newCoreJob(t jobType, deadline time.Time) *coreJob {
	return &coreJob{
		jtype:     t,
		deadline:  deadline,
		errors:    make([]error, 0),
		starttime: time.Now(),
		replchan:  make(chan struct{}),
	}
}

func (c *coreJob) StartTime() time.Time {
	return c.starttime
}

func (c *coreJob) Deadline() time.Time {
	return c.deadline
}

func (c *coreJob) Type() jobType {
	return c.jtype
}

func (c *coreJob) ReplChan() chan struct{} {
	return c.replchan
}

func (c *coreJob) Errors() []error {
	return c.errors
}

type collectorJob struct {
	*coreJob
	collector   collectsMetrics
	metricTypes []core.Metric
	metrics     []core.Metric
}

func newCollectorJob(metricTypes []core.Metric, deadlineDuration time.Duration, collector collectsMetrics) *collectorJob {
	return &collectorJob{
		collector:   collector,
		metricTypes: metricTypes,
		metrics:     []core.Metric{},
		coreJob:     newCoreJob(collectJobType, time.Now().Add(deadlineDuration)),
	}
}

func (c *collectorJob) Run() {
	ret, errs := c.collector.CollectMetrics(c.metricTypes, c.Deadline())
	logger.Debugf("Scheduler.CollectorJob.Run", "We collected: %v err: %v", ret, errs)
	c.metrics = ret
	if errs != nil {
		c.errors = errs
	}
	c.replchan <- struct{}{}
}

type processJob struct {
	*coreJob
	parentJob     job
	metrics       []core.Metric
	pluginName    string
	pluginVersion int
}

func newProcessJob(parentJob job, pluginName string, pluginVersion int) *processJob {
	return &processJob{
		parentJob:     parentJob,
		pluginName:    pluginName,
		pluginVersion: pluginVersion,
		metrics:       []core.Metric{},
		coreJob:       newCoreJob(processJobType, parentJob.Deadline()),
	}
}

type publisherJob struct {
	*coreJob
	parentJob     job
	publisher     publishesMetrics
	pluginName    string
	pluginVersion int
	config        map[string]ctypes.ConfigValue
	contentType   string
}

func newPublishJob(parentJob job, pluginName string, pluginVersion int, contentType string, config map[string]ctypes.ConfigValue, publisher publishesMetrics) *publisherJob {
	return &publisherJob{
		parentJob:     parentJob,
		publisher:     publisher,
		pluginName:    pluginName,
		pluginVersion: pluginVersion,
		coreJob:       newCoreJob(publishJobType, parentJob.Deadline()),
		config:        config,
		contentType:   contentType,
	}
}

func (p *publisherJob) Run() {
	logger.Debugf("Scheduler.PublisherJob.Run", "Starting publish job.")
	logger.Debugf("Scheduler.PublisherJob.Run", "Publishing - contentType: %v pluginName: %v version: %v config: %v", p.contentType, p.pluginName, p.pluginVersion, p.config)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	switch p.parentJob.Type() {
	case collectJobType:
		switch p.contentType {
		case plugin.PulseGOBContentType:
			metrics := make([]plugin.PluginMetricType, len(p.parentJob.(*collectorJob).metrics))
			for i, m := range p.parentJob.(*collectorJob).metrics {
				metrics[i] = *plugin.NewPluginMetricType(m.Namespace(), m.Data())
			}
			enc.Encode(metrics)
		default:
			panic("unsupported content type")
		}
	default:
		panic("unsupported job type")
	}
	logger.Debugf("Scheduler.PublisherJob.Run", "content: %v", buf.Bytes())
	errs := p.publisher.PublishMetrics(p.contentType, buf.Bytes(), p.pluginName, p.pluginVersion, p.config)
	if errs != nil {
		p.errors = append(p.errors, errs...)
	}

	p.replchan <- struct{}{}
}
