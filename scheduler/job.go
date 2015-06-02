package scheduler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	"github.com/intelsdi-x/pulse/pkg/logger"
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
	collector      CollectsMetrics
	metricTypes    []core.RequestedMetric
	metrics        []core.Metric
	configDataTree *cdata.ConfigDataTree
}

func newCollectorJob(metricTypes []core.RequestedMetric, deadlineDuration time.Duration, collector CollectsMetrics, cdt *cdata.ConfigDataTree) job {
	return &collectorJob{
		collector:      collector,
		metricTypes:    metricTypes,
		metrics:        []core.Metric{},
		coreJob:        newCoreJob(collectJobType, time.Now().Add(deadlineDuration)),
		configDataTree: cdt,
	}
}

type metric struct {
	namespace []string
	version   int
	config    *cdata.ConfigDataNode
}

func (m *metric) Namespace() []string {
	return m.namespace
}

func (m *metric) Config() *cdata.ConfigDataNode {
	return m.config
}

func (m *metric) Version() int {
	return m.version
}

func (m *metric) Data() interface{}             { return nil }
func (m *metric) LastAdvertisedTime() time.Time { return time.Unix(0, 0) }

func (c *collectorJob) Run() {
	metrics := make([]core.Metric, len(c.metricTypes))
	for i, rmt := range c.metricTypes {
		metrics[i] = &metric{
			namespace: rmt.Namespace(),
			version:   rmt.Version(),
			config:    c.configDataTree.Get(rmt.Namespace()),
		}
	}
	ret, errs := c.collector.CollectMetrics(metrics, c.Deadline())
	logger.Debugf("Scheduler.CollectorJob.Run", "We collected: %v err: %v", ret, errs)
	c.metrics = ret
	if errs != nil {
		c.errors = errs
	}
	c.replchan <- struct{}{}
}

//todo rename to processJob
type processJob struct {
	*coreJob
	parentJob     job
	metrics       []core.Metric
	pluginName    string
	pluginVersion int
}

func (pr *processJob) Run() {
	// logger.Debugf("Scheduler.ProcessorJob.Run", "Starting processor job.")
	// logger.Debugf("Scheduler.ProcessorJob.Run", "Processor - contentType: %v pluginName: %v version: %v config: %v", p.contentType, p.pluginName, p.pluginVersion, p.config)
	// var buf bytes.Buffer
	// enc := gob.NewEncoder(&buf)

	// switch p.parentJob.Type() {
	// case collectJobType:
	// 	switch p.contentType {
	// 	case plugin.PulseGOBContentType:
	// 		metrics := make([]plugin.PluginMetricType, len(p.parentJob.(*collectorJob).metrics))
	// 		for i, m := range p.parentJob.(*collectorJob).metrics {
	// 			metrics[i] = *plugin.NewPluginMetricType(m.Namespace(), m.Data())
	// 		}
	// 		enc.Encode(metrics)
	// 	default:
	// 		panic(fmt.Sprintf("unsupported content type. {plugin name: %s version: %v content-type: '%v'}", p.pluginName, p.pluginVersion, p.contentType))
	// 	}
	// default:
	// 	panic("unsupported job type")
	// }
	// logger.Debugf("Scheduler.ProcessorJob.Run", "content: %v", buf.Bytes())
	// errs := p.publisher.PublishMetrics(p.contentType, buf.Bytes(), p.pluginName, p.pluginVersion, p.config)
	// if errs != nil {
	// 	p.errors = append(p.errors, errs...)
	// }

	// p.replchan <- struct{}{}
}

func newProcessJob(parentJob job, pluginName string, pluginVersion int) job {
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
	publisher     PublishesMetrics
	pluginName    string
	pluginVersion int
	config        map[string]ctypes.ConfigValue
	contentType   string
}

func newPublishJob(parentJob job, pluginName string, pluginVersion int, contentType string, config map[string]ctypes.ConfigValue, publisher PublishesMetrics) job {
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
			panic(fmt.Sprintf("unsupported content type. {plugin name: %s version: %v content-type: '%v'}", p.pluginName, p.pluginVersion, p.contentType))
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
