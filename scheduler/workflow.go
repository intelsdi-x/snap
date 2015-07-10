package scheduler

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/gomit"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/scheduler_event"
	"github.com/intelsdi-x/pulse/scheduler/wmap"
)

type WorkflowState int

const (
	WorkflowStopped WorkflowState = iota
	WorkflowStarted
)

var (
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
	err := convertCollectionNode(wfMap.CollectNode, wf)
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
		wf.metrics[i] = m
	}

	// Get our config data tree
	cdt, err := cnode.GetConfigTree()
	if err != nil {
		return err
	}
	wf.configTree = cdt
	// Iterate over first level process nodes
	pr, err := convertProcessNode(cnode.ProcessNodes)
	if err != nil {
		return err
	}
	wf.processNodes = pr
	// Iterate over first level publish nodes
	pu, err := convertPublishNode(cnode.PublishNodes)
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
		prC, err := convertProcessNode(p.ProcessNodes)
		if err != nil {
			return nil, err
		}
		puC, err := convertPublishNode(p.PublishNodes)
		if err != nil {
			return nil, err
		}

		// If version is not 1+ we use -1 to indicate we want
		// the plugin manager to select the highest version
		// available on plugin calls
		if p.Version < 1 {
			p.Version = -1
		}
		prNodes[i] = &processNode{
			name:         p.Name,
			version:      p.Version,
			config:       cdn,
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
		if p.Version < 1 {
			p.Version = -1
		}
		puNodes[i] = &publishNode{
			name:    p.Name,
			version: p.Version,
			config:  cdn,
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
}

type processNode struct {
	name               string
	version            int
	config             *cdata.ConfigDataNode
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

// BindPluginContentTypes
func (s *schedulerWorkflow) BindPluginContentTypes(mm managesPluginContentTypes) error {
	bindPluginContentTypes(s.publishNodes, s.processNodes, mm, []string{plugin.PulseGOBContentType})
	return nil
}

func bindPluginContentTypes(pus []*publishNode, prs []*processNode, mm managesPluginContentTypes, lct []string) error {
	for _, pr := range prs {
		act, rct, err := mm.GetPluginContentTypes(pr.Name(), core.ProcessorPluginType, pr.Version())
		if err != nil {
			return err
		}

		for _, ac := range act {
			for _, lc := range lct {
				// if the return contenet type from the previous node matches
				// the accept content type for this node set it as the
				// inbound content type
				if ac == lc {
					pr.InboundContentType = ac
				}
			}
		}
		// if the inbound content type isn't set yet pulse may be able to do
		// the conversion
		if pr.InboundContentType == "" {
			for _, ac := range act {
				switch ac {
				case plugin.PulseGOBContentType:
					pr.InboundContentType = plugin.PulseGOBContentType
				case plugin.PulseJSONContentType:
					pr.InboundContentType = plugin.PulseJSONContentType
				case plugin.PulseAllContentType:
					pr.InboundContentType = plugin.PulseGOBContentType
				}
			}
			// else we return an error
			if pr.InboundContentType == "" {
				return fmt.Errorf("Invalid workflow.  Plugin '%s' does not accept the pulse content types or the types '%v' returned from the previous node.", pr.Name(), lct)
			}
		}
		//continue the walk down the nodes
		bindPluginContentTypes(pr.PublishNodes, pr.ProcessNodes, mm, rct)
	}
	for _, pu := range pus {
		act, _, err := mm.GetPluginContentTypes(pu.Name(), core.PublisherPluginType, pu.Version())
		if err != nil {
			return err
		}
		// if the inbound content type isn't set yet pulse may be able to do
		// the conversion
		if pu.InboundContentType == "" {
			for _, ac := range act {
				switch ac {
				case plugin.PulseGOBContentType:
					pu.InboundContentType = plugin.PulseGOBContentType
				case plugin.PulseJSONContentType:
					pu.InboundContentType = plugin.PulseJSONContentType
				case plugin.PulseAllContentType:
					pu.InboundContentType = plugin.PulseGOBContentType
				}
			}
			// else we return an error
			if pu.InboundContentType == "" {
				return fmt.Errorf("Invalid workflow.  Plugin '%s' does not accept the pulse content types or the types '%v' returned from the previous node.", pu.Name(), lct)
			}
		}
	}
	return nil
}

// Start starts a workflow
func (s *schedulerWorkflow) Start(t *task) {
	s.state = WorkflowStarted
	j := newCollectorJob(s.metrics, t.deadlineDuration, t.metricsManager, t.workflow.configTree)

	// dispatch 'collect' job to be worked
	j = t.manager.Work(j)
	if len(j.Errors()) != 0 {

		t.failedRuns++
		t.lastFailureTime = t.lastFireTime
		t.lastFailureMessage = j.Errors()[len(j.Errors())-1].Error()
		event := new(scheduler_event.MetricCollectionFailedEvent)
		event.TaskID = t.id
		event.Errors = j.Errors()
		defer s.eventEmitter.Emit(event)
		return
	}

	// Send event
	event := new(scheduler_event.MetricCollectedEvent)
	event.TaskID = t.id
	event.Metrics = j.(*collectorJob).metrics
	defer s.eventEmitter.Emit(event)

	// walk through the tree and dispatch work
	s.workJobs(s.processNodes, s.publishNodes, t, j)
}

func (s *schedulerWorkflow) State() WorkflowState {
	return s.state
}

func (s *schedulerWorkflow) StateString() string {
	return WorkflowStateLookup[s.state]
}

func (s *schedulerWorkflow) workJobs(prs []*processNode, pus []*publishNode, t *task, pj job) {
	for _, pr := range prs {
		j := newProcessJob(pj, pr.Name(), pr.Version(), pr.InboundContentType, pr.config.Table(), t.metricsManager)
		j = t.manager.Work(j)
		if len(j.Errors()) != 0 {
			t.failedRuns++
			t.lastFailureTime = t.lastFireTime
			t.lastFailureMessage = j.Errors()[len(j.Errors())-1].Error()
			return
		}

		s.workJobs(pr.ProcessNodes, pr.PublishNodes, t, j)
	}
	for _, pu := range pus {
		j := newPublishJob(pj, pu.Name(), pu.Version(), pu.InboundContentType, pu.config.Table(), t.metricsManager)
		j = t.manager.Work(j)
		if len(j.Errors()) != 0 {
			t.failedRuns++
			t.lastFailureTime = t.lastFireTime
			t.lastFailureMessage = j.Errors()[len(j.Errors())-1].Error()
			return
		}
	}
}
