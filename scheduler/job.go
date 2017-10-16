/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/intelsdi-x/snap/pkg/promise"
)

const (
	collectJobType jobType = iota
	publishJobType
	processJobType
)

const (
	defaultDeadline = time.Duration(5 * time.Second)
)

// Represents a queued job, together with a synchronization
// barrier to signal job completion (successful or otherwise).
//
// Functions that operate on this type (IsComplete, Complete,
// Await) are idempotent and thread-safe.
type queuedJob interface {
	Job() job
	Promise() Promise
}

type qj struct {
	job     job
	promise Promise
}

func newQueuedJob(job job) queuedJob {
	return &qj{
		job:     job,
		promise: NewPromise(),
	}
}

// Returns the underlying job.
func (j *qj) Job() job {
	return j.job
}

// Returns the underlying promise.
func (j *qj) Promise() Promise {
	return j.promise
}

// Primary type for job inside
// the scheduler.  Job encompasses all
// all job types -- collect, process, and publish.
type job interface {
	AddErrors(errs ...error)
	Errors() []error
	StartTime() time.Time
	Deadline() time.Time
	Name() string
	Version() int
	Type() jobType
	TypeString() string
	TaskID() string
	Run()
	Metrics() []core.Metric
}

type jobType int

type coreJob struct {
	sync.Mutex
	name      string
	version   int
	taskID    string
	jtype     jobType
	deadline  time.Time
	starttime time.Time
	errors    []error
}

func newCoreJob(t jobType, deadline time.Time, taskID string, name string, version int) *coreJob {
	return &coreJob{
		jtype:     t,
		name:      name,
		version:   version,
		deadline:  deadline,
		taskID:    taskID,
		errors:    make([]error, 0),
		starttime: time.Now(),
	}
}

func (c *coreJob) StartTime() time.Time {
	return c.starttime
}

func (c *coreJob) Deadline() time.Time {
	return c.deadline
}

func (c *coreJob) Name() string {
	return c.name
}

func (c *coreJob) Version() int {
	return c.version
}

func (c *coreJob) Type() jobType {
	return c.jtype
}

func (c *coreJob) TypeString() string {
	switch c.jtype {
	case collectJobType:
		return "collector"

	case processJobType:
		return "processor"

	case publishJobType:
		return "publisher"
	}
	return "unknown"
}

func (c *coreJob) AddErrors(errs ...error) {
	c.Lock()
	defer c.Unlock()
	c.errors = append(c.errors, errs...)
}

func (c *coreJob) Errors() []error {
	return c.errors
}

func (c *coreJob) TaskID() string {
	return c.taskID
}

type collectorJob struct {
	*coreJob
	collector      collectsMetrics
	metricTypes    []core.RequestedMetric
	metrics        []core.Metric
	configDataTree *cdata.ConfigDataTree
	tags           map[string]map[string]string
}

func newCollectorJob(
	metricTypes []core.RequestedMetric,
	deadlineDuration time.Duration,
	collector collectsMetrics,
	cdt *cdata.ConfigDataTree,
	taskID string,
	tags map[string]map[string]string,
) job {
	return &collectorJob{
		collector:      collector,
		metricTypes:    metricTypes,
		metrics:        []core.Metric{},
		coreJob:        newCoreJob(collectJobType, time.Now().Add(deadlineDuration), taskID, "", 0),
		configDataTree: cdt,
		tags:           tags,
	}
}

type metric struct {
	namespace core.Namespace
	version   int
	config    *cdata.ConfigDataNode
}

func (m *metric) Namespace() core.Namespace {
	return m.namespace
}

func (m *metric) Config() *cdata.ConfigDataNode {
	return m.config
}

func (m *metric) Version() int {
	return m.version
}

func (m *metric) Data() interface{}             { return nil }
func (m *metric) Description() string           { return "" }
func (m *metric) Unit() string                  { return "" }
func (m *metric) Tags() map[string]string       { return nil }
func (m *metric) LastAdvertisedTime() time.Time { return time.Unix(0, 0) }
func (m *metric) Timestamp() time.Time          { return time.Unix(0, 0) }

func (c *collectorJob) Metrics() []core.Metric {
	return c.metrics
}

func (c *collectorJob) Run() {
	log.WithFields(log.Fields{
		"_module":      "scheduler-job",
		"block":        "run",
		"job-type":     "collector",
		"metric-count": len(c.metricTypes),
	}).Debug("starting collector job")

	for ns, tags := range c.tags {
		for k, v := range tags {
			log.WithFields(log.Fields{
				"_module":  "scheduler-job",
				"block":    "run",
				"job-type": "collector",
				"ns":       ns,
				"tag-key":  k,
				"tag-val":  v,
			}).Debug("Tags sent to collectorJob")
		}
	}

	ret, errs := c.collector.CollectMetrics(c.TaskID(), c.tags)

	log.WithFields(log.Fields{
		"_module":      "scheduler-job",
		"block":        "run",
		"job-type":     "collector",
		"metric-count": len(ret),
	}).Debug("collector run completed")

	c.metrics = ret
	if errs != nil {
		for _, e := range errs {
			log.WithFields(log.Fields{
				"_module":  "scheduler-job",
				"block":    "run",
				"job-type": "collector",
				"error":    e,
			}).Error("collector run error")
		}
		c.AddErrors(errs...)
	}
}

type processJob struct {
	*coreJob
	processor processesMetrics
	parentJob job
	metrics   []core.Metric
	config    map[string]ctypes.ConfigValue
}

func (pr *processJob) Metrics() []core.Metric {
	return pr.metrics
}

func newProcessJob(parentJob job, pluginName string, pluginVersion int, contentType string, config map[string]ctypes.ConfigValue, processor processesMetrics, taskID string) job {
	return &processJob{
		parentJob: parentJob,
		metrics:   []core.Metric{},
		coreJob:   newCoreJob(processJobType, parentJob.Deadline(), taskID, pluginName, pluginVersion),
		config:    config,
		processor: processor,
	}
}

func (p *processJob) Run() {
	log.WithFields(log.Fields{
		"_module":        "scheduler-job",
		"block":          "run",
		"job-type":       "processor",
		"plugin-name":    p.name,
		"plugin-version": p.version,
		"plugin-config":  p.config,
	}).Debug("starting processor job")

	mts, errs := p.processor.ProcessMetrics(p.parentJob.Metrics(), p.config, p.taskID, p.name, p.version)
	if errs != nil {
		for _, e := range errs {
			log.WithFields(log.Fields{
				"_module":        "scheduler-job",
				"block":          "run",
				"job-type":       "processor",
				"plugin-name":    p.name,
				"plugin-version": p.version,
				"plugin-config":  p.config,
				"error":          e.Error(),
			}).Error("error with processor job")
		}
		p.AddErrors(errs...)
	}
	p.metrics = mts
}

type publisherJob struct {
	*coreJob
	parentJob job
	publisher publishesMetrics
	config    map[string]ctypes.ConfigValue
}

func (pu *publisherJob) Metrics() []core.Metric {
	return []core.Metric{}
}

func newPublishJob(parentJob job, pluginName string, pluginVersion int, contentType string, config map[string]ctypes.ConfigValue, publisher publishesMetrics, taskID string) job {
	return &publisherJob{
		parentJob: parentJob,
		publisher: publisher,
		coreJob:   newCoreJob(publishJobType, parentJob.Deadline(), taskID, pluginName, pluginVersion),
		config:    config,
	}
}

func (p *publisherJob) Run() {
	log.WithFields(log.Fields{
		"_module":        "scheduler-job",
		"block":          "run",
		"job-type":       "publisher",
		"plugin-name":    p.name,
		"plugin-version": p.version,
		"plugin-config":  p.config,
	}).Debug("starting publisher job")

	errs := p.publisher.PublishMetrics(p.parentJob.Metrics(), p.config, p.taskID, p.name, p.version)
	if errs != nil {
		for _, e := range errs {
			log.WithFields(log.Fields{
				"_module":        "scheduler-job",
				"block":          "run",
				"job-type":       "publisher",
				"plugin-name":    p.name,
				"plugin-version": p.version,
				"plugin-config":  p.config,
				"error":          e.Error(),
			}).Error("error with publisher job")
		}
		p.AddErrors(errs...)
	}
}
