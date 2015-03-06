package schedule

import (
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

const (
	//Task states
	TaskStopped TaskState = iota
	TaskSpinning
	TaskFiring
)

type metricType struct {
	config                  *cdata.ConfigDataNode
	namespace               []string
	lastAdvertisedTimestamp int64
	version                 int
}

func (m *metricType) Version() int {
	return m.version
}

func (m *metricType) Namespace() []string {
	return m.namespace
}

func (m *metricType) LastAdvertisedTimestamp() int64 {
	return m.lastAdvertisedTimestamp
}

func (m *metricType) Config() *cdata.ConfigDataNode {
	return m.config
}

func newMetricType(mt core.MetricType, config *cdata.ConfigDataNode) *metricType {
	return &metricType{
		namespace:               mt.Namespace(),
		version:                 mt.Version(),
		lastAdvertisedTimestamp: mt.LastAdvertisedTimestamp(),
		config:                  config,
	}
}

type Task struct {
	schResponseChan chan ScheduleResponse
	killChan        chan struct{}
	schedule        Schedule
	workflow        Workflow
	metricTypes     []*metricType
	mu              sync.Mutex //protects state
	state           TaskState
	creationTime    time.Time
	lastFireTime    time.Time
}

type TaskState int

func NewTask(s Schedule, mtc []*metricType) *Task {
	return &Task{
		schResponseChan: make(chan ScheduleResponse),
		killChan:        make(chan struct{}),
		metricTypes:     mtc,
		schedule:        s,
		state:           TaskStopped,
		creationTime:    time.Now(),
		workflow:        NewWorkflow(),
	}
}

func (t *Task) Spin() {
	// We need to lock long enough to change state
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == TaskStopped {
		t.state = TaskSpinning
		// spin in a goroutine
		go t.spin()
	}
}

func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state != TaskStopped {
		t.killChan <- struct{}{}
	}
}

func (t *Task) spin() {
	for {
		// Start go routine to wait on schedule
		go t.waitForSchedule()
		// wait here on
		//  schResponseChan - response from schedule
		//  killChan - signals task needs to be stopped
		select {
		case sr := <-t.schResponseChan:
			// If response show this schedule is stil active we fire
			if sr.State() == ScheduleActive {
				t.lastFireTime = time.Now()
				t.fire()
			}
			// TODO stop task on schedule error state or end state
		case <-t.killChan:
			// Only here can it truly be stopped
			t.state = TaskStopped
			break
		}
	}
}

func (t *Task) fire() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.state = TaskFiring
	t.workflow.Start(t, WorkManager)
	t.state = TaskSpinning
}

func (t *Task) waitForSchedule() {
	t.schResponseChan <- t.schedule.Wait(t.lastFireTime)
}
