package scheduler

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/pkg/logger"
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

	NullCollectNodeError        = errors.New("Missing collection node in workflow map")
	NoMetricsInCollectNodeError = errors.New("Collection node has not metrics defined to collect")
)

// WmapToWorkflow attempts to convert a wmap.WorkflowMap to a schedulerWorkflow instance.
func wmapToWorkflow(wfMap *wmap.WorkflowMap) (*schedulerWorkflow, error) {
	fmt.Println("- WORKFLOW - ")
	defer fmt.Println("- WORKFLOW - ")
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
		return NullCollectNodeError
	}
	// Collection node has at least one metric in it
	if len(cnode.Metrics) < 1 {
		return NoMetricsInCollectNodeError
	}
	// Get core.RequestedMetric metrics
	wf.metrics = cnode.GetRequestedMetrics()
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
			Name:         p.Name,
			Version:      p.Version,
			Config:       cdn,
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
			Name:    p.Name,
			Version: p.Version,
			Config:  cdn,
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
	workflowMap *wmap.WorkflowMap
}

type processNode struct {
	Name               string
	Version            int
	Config             *cdata.ConfigDataNode
	ProcessNodes       []*processNode
	PublishNodes       []*publishNode
	InboundContentType string
}

type publishNode struct {
	Name               string
	Version            int
	Config             *cdata.ConfigDataNode
	InboundContentType string
}

type wfContentTypes map[string]map[string][]string

// BindPluginContentTypes uses the provided ManagesMetrics to
func (s *schedulerWorkflow) BindPluginContentTypes(mm ManagesPluginContentTypes) error {
	// Default returned types from a collection
	// possibleContentTypes := []string{"pulse.*", "pulse.gob", "pulse.json"}
	// Walk nodes to query and bind the content types required
	// top level
	logger.SetLevel(logger.DebugLevel)
	bindPluginContentTypes(s.publishNodes, s.processNodes, mm, []string{plugin.PulseGOBContentType})
	return nil
}

func bindPluginContentTypes(pus []*publishNode, prs []*processNode, mm ManagesPluginContentTypes, lct []string) error {
	//todo lastcontentTypes needs to be passed in
	for _, pr := range prs {
		act, rct, err := mm.GetPluginContentTypes(pr.Name, core.ProcessorPluginType, pr.Version)
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
				return fmt.Errorf("Invalid workflow.  Plugin '%s' does not accept the pulse content types or the types '%v' returned from the previous node.", pr.Name, lct)
			}
		}
		//continue the walk down the nodes
		bindPluginContentTypes(pr.PublishNodes, pr.ProcessNodes, mm, rct)
	}
	for _, pu := range pus {
		act, _, err := mm.GetPluginContentTypes(pu.Name, core.PublisherPluginType, pu.Version)
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
				return fmt.Errorf("Invalid workflow.  Plugin '%s' does not accept the pulse content types or the types '%v' returned from the previous node.", pu.Name, lct)
			}
		}
	}
	return nil
}

// Start starts a workflow
func (w *schedulerWorkflow) Start(t *task) {
	w.state = WorkflowStarted
	j := newCollectorJob(w.metrics, t.deadlineDuration, t.metricsManager, t.workflow.configTree)

	// dispatch 'collect' job to be worked
	j = t.manager.Work(j)

	// walk through the tree and dispatch work
	w.workJobs(w.processNodes, w.publishNodes, t, j)
}

func (s *schedulerWorkflow) State() WorkflowState {
	return s.state
}

func (s *schedulerWorkflow) StateString() string {
	return WorkflowStateLookup[s.state]
}

//TODO JC pass in manages work instead of task
func (w *schedulerWorkflow) workJobs(prs []*processNode, pus []*publishNode, t *task, pj job) {
	for _, pr := range prs {
		j := newProcessJob(pj, pr.Name, pr.Version)
		j = t.manager.Work(j)
		for _, npr := range pr.ProcessNodes {
			w.workJobs(npr.ProcessNodes, npr.PublishNodes, t, j)
		}
	}
	for _, pu := range pus {
		j := newPublishJob(pj, pu.Name, pu.Version, pu.InboundContentType, pu.Config.Table(), t.metricsManager)
		j = t.manager.Work(j)
	}
}
