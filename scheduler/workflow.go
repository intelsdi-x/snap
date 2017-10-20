/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/intelsdi-x/gomit"
	log "github.com/sirupsen/logrus"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/scheduler_event"
	"github.com/intelsdi-x/snap/scheduler/wmap"
)

// WorkflowState int type
type WorkflowState int

// Workflow state constants
const (
	WorkflowStopped WorkflowState = iota
	WorkflowStarted
)

// WorkflowStateLookup map and error vars
var (
	workflowLogger = schedulerLogger.WithField("_module", "scheduler-workflow")

	WorkflowStateLookup = map[WorkflowState]string{
		WorkflowStopped: "Stopped",
		WorkflowStarted: "Started",
	}

	ErrNullCollectNode        = errors.New("Missing collection node in workflow map")
	ErrNoMetricsInCollectNode = errors.New("Collection node has not metrics defined to collect")
)

// WmapToWorkflow attempts to convert a wmap.WorkflowMap to a schedulerWorkflow instance.
func wmapToWorkflow(wfMap *wmap.WorkflowMap) (*schedulerWorkflow, error) {
	wf := &schedulerWorkflow{}
	err := convertCollectionNode(wfMap.Collect, wf)
	if err != nil {
		return nil, err
	}
	// ***
	// TODO validate workflow makes sense here
	// - flows that don't end in publishers?
	// - duplicate child nodes anywhere?
	//***
	// Retain a copy of the original workflow map
	wf.workflowMap = wfMap
	return wf, nil
}

func convertCollectionNode(cnode *wmap.CollectWorkflowMapNode, wf *schedulerWorkflow) error {
	// Collection root
	// Validate collection node exists
	if cnode == nil {
		return ErrNullCollectNode
	}
	// Collection node has at least one metric in it
	if len(cnode.Metrics) < 1 {
		return ErrNoMetricsInCollectNode
	}
	// Get core.RequestedMetric metrics
	mts := cnode.GetMetrics()
	wf.metrics = make([]core.RequestedMetric, len(mts))
	for i, m := range mts {
		wf.metrics[i] = &metric{namespace: core.NewNamespace(m.Namespace()...), version: m.Version()}
	}
	// get tags defined
	wf.tags = cnode.GetTags()

	// Get our config data tree
	cdt, err := cnode.GetConfigTree()
	if err != nil {
		return err
	}
	wf.configTree = cdt
	// Iterate over first level process nodes
	pr, err := convertProcessNode(cnode.Process)
	if err != nil {
		return err
	}
	wf.processNodes = pr
	// Iterate over first level publish nodes
	pu, err := convertPublishNode(cnode.Publish)
	if err != nil {
		return err
	}
	wf.publishNodes = pu
	return nil
}

func convertProcessNode(pr []wmap.ProcessWorkflowMapNode) ([]*processNode, error) {
	prNodes := make([]*processNode, len(pr))
	for i, p := range pr {
		cdn, err := p.GetConfigNode()
		if err != nil {
			return nil, err
		}
		prC, err := convertProcessNode(p.Process)
		if err != nil {
			return nil, err
		}
		puC, err := convertPublishNode(p.Publish)
		if err != nil {
			return nil, err
		}

		// If version is not 1+ we use -1 to indicate we want
		// the plugin manager to select the highest version
		// available on plugin calls
		if p.PluginVersion < 1 {
			p.PluginVersion = -1
		}
		p.PluginName = strings.ToLower(p.PluginName)
		prNodes[i] = &processNode{
			name:         p.PluginName,
			version:      p.PluginVersion,
			config:       cdn,
			Target:       p.Target,
			ProcessNodes: prC,
			PublishNodes: puC,
		}
	}
	return prNodes, nil
}

func convertPublishNode(pu []wmap.PublishWorkflowMapNode) ([]*publishNode, error) {
	puNodes := make([]*publishNode, len(pu))
	for i, p := range pu {

		cdn, err := p.GetConfigNode()
		if err != nil {
			return nil, err
		}
		// If version is not 1+ we use -1 to indicate we want
		// the plugin manager to select the highest version
		// available on plugin calls
		if p.PluginVersion < 1 {
			p.PluginVersion = -1
		}
		p.PluginName = strings.ToLower(p.PluginName)
		puNodes[i] = &publishNode{
			name:    p.PluginName,
			version: p.PluginVersion,
			config:  cdn,
			Target:  p.Target,
		}
	}
	return puNodes, nil
}

type schedulerWorkflow struct {
	state WorkflowState
	// Metrics to collect
	metrics []core.RequestedMetric
	// The config data tree for collectors
	configTree   *cdata.ConfigDataTree
	processNodes []*processNode
	publishNodes []*publishNode
	// workflowMap used to generate this workflow
	workflowMap  *wmap.WorkflowMap
	eventEmitter gomit.Emitter
	tags         map[string]map[string]string
}

type processNode struct {
	name               string
	version            int
	config             *cdata.ConfigDataNode
	Target             string
	ProcessNodes       []*processNode
	PublishNodes       []*publishNode
	InboundContentType string
}

func (p *processNode) Name() string {
	return p.name
}

func (p *processNode) Version() int {
	return p.version
}

func (p *processNode) Config() *cdata.ConfigDataNode {
	return p.config
}

func (p *processNode) TypeName() string {
	return "processor"
}

type publishNode struct {
	name               string
	version            int
	config             *cdata.ConfigDataNode
	Target             string
	InboundContentType string
}

func (p *publishNode) Name() string {
	return p.name
}

func (p *publishNode) Version() int {
	return p.version
}

func (p *publishNode) Config() *cdata.ConfigDataNode {
	return p.config
}

func (p *publishNode) TypeName() string {
	return "publisher"
}

type wfContentTypes map[string]map[string][]string

// Start starts a workflow
func (s *schedulerWorkflow) Start(t *task) {
	workflowLogger.WithFields(log.Fields{
		"_block":    "workflow-start",
		"task-id":   t.id,
		"task-name": t.name,
	}).Debug("Starting workflow")
	s.state = WorkflowStarted
	j := newCollectorJob(s.metrics, t.deadlineDuration, t.metricsManager, t.workflow.configTree, t.id, s.tags)

	// dispatch 'collect' job to be worked
	// Block until the job has been either run or skipped.
	errors := t.manager.Work(j).Promise().Await()

	if len(errors) > 0 {
		t.RecordFailure(errors)
		event := new(scheduler_event.MetricCollectionFailedEvent)
		event.TaskID = t.id
		event.Errors = errors
		defer s.eventEmitter.Emit(event)
		return
	}

	// Send event
	event := new(scheduler_event.MetricCollectedEvent)
	event.TaskID = t.id
	event.Metrics = j.(*collectorJob).metrics
	defer s.eventEmitter.Emit(event)

	// walk through the tree and dispatch work
	workJobs(s.processNodes, s.publishNodes, t, j)
}

func (s *schedulerWorkflow) State() WorkflowState {
	return s.state
}

func (s *schedulerWorkflow) StateString() string {
	return WorkflowStateLookup[s.state]
}

func (s *schedulerWorkflow) StreamStart(t *task, metrics []core.Metric) {
	j := &collectorJob{
		collector:      t.metricsManager,
		metricTypes:    []core.RequestedMetric{},
		metrics:        metrics,
		coreJob:        newCoreJob(collectJobType, time.Now().Add(t.deadlineDuration), t.id, "", 0),
		configDataTree: t.workflow.configTree,
		tags:           t.workflow.tags,
	}
	// Send event
	event := new(scheduler_event.MetricCollectedEvent)
	event.TaskID = t.id
	event.Metrics = j.metrics
	defer s.eventEmitter.Emit(event)
	workJobs(s.processNodes, s.publishNodes, t, j)
}

// workJobs takes a slice of process and publish nodes and submits jobs for each for a task.
// It then iterates down any process nodes to submit their child node jobs for the task
func workJobs(prs []*processNode, pus []*publishNode, t *task, pj job) {
	// optimize for no jobs
	if len(prs) == 0 && len(pus) == 0 {
		return
	}
	// Create waitgroup to block until all jobs are submitted
	wg := &sync.WaitGroup{}
	workflowLogger.WithFields(log.Fields{
		"_block":              "work-jobs",
		"task-id":             t.id,
		"task-name":           t.name,
		"count-process-nodes": len(prs),
		"count-publish-nodes": len(pus),
		"parent-node-type":    pj.TypeString(),
	}).Debug("Batch submission of process and publish nodes")
	// range over the process jobs and call submitProcessJob
	for _, pr := range prs {
		// increment the wait group (before starting goroutine to prevent a race condition)
		wg.Add(1)
		// Start goroutine to submit the process job
		go submitProcessJob(pj, t, wg, pr)
	}
	// range over the publish jobs and call submitPublishJob
	for _, pu := range pus {
		// increment the wait group (before starting goroutine to prevent a race condition)
		wg.Add(1)
		// Start goroutine to submit the process job
		go submitPublishJob(pj, t, wg, pu)
	}
	// Wait until all job submisson goroutines are done
	wg.Wait()
	workflowLogger.WithFields(log.Fields{
		"_block":              "work-jobs",
		"task-id":             t.id,
		"task-name":           t.name,
		"count-process-nodes": len(prs),
		"count-publish-nodes": len(pus),
		"parent-node-type":    pj.TypeString(),
	}).Debug("Batch submission complete")
}

func submitProcessJob(pj job, t *task, wg *sync.WaitGroup, pr *processNode) {
	// Decrement the waitgroup
	defer wg.Done()
	// Create a new process job
	mgr, err := t.RemoteManagers.Get(pr.Target)
	if err != nil {
		t.RecordFailure([]error{err})
		workflowLogger.WithFields(log.Fields{
			"_block":           "submit-prblish-job",
			"task-id":          t.id,
			"task-name":        t.name,
			"prblish-name":     pr.Name(),
			"prblish-version":  pr.Version(),
			"parent-node-type": pj.TypeString(),
		}).Warn("Error getting control instance")
		return
	}
	j := newProcessJob(pj, pr.Name(), pr.Version(), pr.InboundContentType, pr.config.Table(), mgr, t.id)
	workflowLogger.WithFields(log.Fields{
		"_block":           "submit-process-job",
		"task-id":          t.id,
		"task-name":        t.name,
		"process-name":     pr.Name(),
		"process-version":  pr.Version(),
		"parent-node-type": pj.TypeString(),
	}).Debug("Submitting process job")
	// Submit the job against the task.managesWork
	errors := t.manager.Work(j).Promise().Await()
	// Check for errors and update the task
	if len(errors) != 0 {
		// Record the failures in the task
		// note: this function is thread safe against t
		t.RecordFailure(errors)
		workflowLogger.WithFields(log.Fields{
			"_block":           "submit-process-job",
			"task-id":          t.id,
			"task-name":        t.name,
			"process-name":     pr.Name(),
			"process-version":  pr.Version(),
			"parent-node-type": pj.TypeString(),
		}).Warn("Process job failed")
		return
	}
	workflowLogger.WithFields(log.Fields{
		"_block":           "submit-process-job",
		"task-id":          t.id,
		"task-name":        t.name,
		"process-name":     pr.Name(),
		"process-version":  pr.Version(),
		"parent-node-type": pj.TypeString(),
	}).Debug("Process job completed")
	// Iterate into any child process or publish nodes
	workJobs(pr.ProcessNodes, pr.PublishNodes, t, j)
}

func submitPublishJob(pj job, t *task, wg *sync.WaitGroup, pu *publishNode) {
	// Decrement the waitgroup
	defer wg.Done()
	// Create a new process job
	mgr, err := t.RemoteManagers.Get(pu.Target)
	if err != nil {
		t.RecordFailure([]error{err})
		workflowLogger.WithFields(log.Fields{
			"_block":           "submit-publish-job",
			"task-id":          t.id,
			"task-name":        t.name,
			"publish-name":     pu.Name(),
			"publish-version":  pu.Version(),
			"parent-node-type": pj.TypeString(),
		}).Warn("Error getting control instance")
		return
	}
	j := newPublishJob(pj, pu.Name(), pu.Version(), pu.InboundContentType, pu.config.Table(), mgr, t.id)
	workflowLogger.WithFields(log.Fields{
		"_block":           "submit-publish-job",
		"task-id":          t.id,
		"task-name":        t.name,
		"publish-name":     pu.Name(),
		"publish-version":  pu.Version(),
		"parent-node-type": pj.TypeString(),
	}).Debug("Submitting publish job")
	// Submit the job against the task.managesWork
	errors := t.manager.Work(j).Promise().Await()
	// Check for errors and update the task
	if len(errors) != 0 {
		// Record the failures in the task
		// note: this function is thread safe against t
		t.RecordFailure(errors)
		workflowLogger.WithFields(log.Fields{
			"_block":           "submit-publish-job",
			"task-id":          t.id,
			"task-name":        t.name,
			"publish-name":     pu.Name(),
			"publish-version":  pu.Version(),
			"parent-node-type": pj.TypeString(),
		}).Warn("Publish job failed")
		return
	}
	workflowLogger.WithFields(log.Fields{
		"_block":           "submit-publish-job",
		"task-id":          t.id,
		"task-name":        t.name,
		"publish-name":     pu.Name(),
		"publish-version":  pu.Version(),
		"parent-node-type": pj.TypeString(),
	}).Debug("Publish job completed")
	// Publish nodes cannot contain child nodes (publish is a terminal node)
	// so unlike process nodes there is not a call to workJobs here for child nodes.
}
