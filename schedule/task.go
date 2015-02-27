package schedule

import (
	"sync"

	"github.com/intelsdilabs/pulse/core"
)

type Task struct {
	killChan    chan struct{}
	schedule    Schedule
	workflow    Workflow
	metricTypes []core.MetricType
	mu          sync.Mutex //protects state
	state       TaskState
}

type TaskState int

const (
	//Task states
	TaskStopped TaskState = iota - 1
	TaskSpinning
	TaskFiring
)

func NewTask(s Schedule) *Task {
	return &Task{
		killChan: make(chan struct{}),
		schedule: s,
		state:    TaskStopped,
		workflow: NewWorkflow(),
	}
}

func (t *Task) MetricTypes() []core.MetricType {
	return t.metricTypes
}

func (t *Task) Spin() {
	if t.state == TaskStopped {
		t.state = TaskSpinning
		go func(kc <-chan struct{}) {
			for {
				select {
				case <-t.schedule.Wait():
					t.fire()
				case <-kc:
					break
				}
			}
		}(t.killChan)
	}
}

func (t *Task) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == TaskStopped {
		return
	}
	t.killChan <- struct{}{}
	t.state = TaskStopped
}

func (t *Task) fire() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state == TaskFiring {
		return
	}
	t.state = TaskFiring

	//routine fires to get work done (and waits and then updates state)
	go func() {
		t.workflow.Start(t, WorkDispatcher)
		t.mu.Lock()
		if t.state == TaskFiring {
			t.state = TaskSpinning
		}
		t.mu.Unlock()
	}()

}
