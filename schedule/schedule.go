package schedule

import (
	"time"

	"github.com/intelsdilabs/pulse/control"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
)

const (
	ScheduleActive ScheduleState = iota
	ScheduleEnded
	ScheduleError
)

var metricManager ManagesMetric

type Schedule interface {
	Wait(time.Time) ScheduleResponse
	Validate() error
}

type ScheduleState int

type ScheduleResponse interface {
	State() ScheduleState
	Error() error
	MissedIntervals() int
}

type ManagesMetric interface {
	SubscribeMetric(metric []string, ver int, cd *cdata.ConfigDataNode) (*cdata.ConfigDataNode, control.SubscriptionError)
	UnsubscribeMetric(metric []string, ver int)
}

type TaskErrors interface {
	Errors() []error
}

type taskErrors struct {
	errs []error
}

func (t *taskErrors) Errors() []error {
	return t.errs
}

type scheduler struct {
}

func (scheduler *scheduler) CreateTask(mt []core.MetricType, s Schedule, cd *cdata.ConfigDataNode) (*Task, TaskErrors) {
	te := &taskErrors{
		errs: make([]error, 0),
	}
	//map MetricType to ConfigDataNode
	mtc := make([]*metricType, 0) //make(map[core.MetricType]*cdata.ConfigDataNode)

	//validate Schedule
	if err := s.Validate(); err != nil {
		te.errs = append(te.errs, err)
		return nil, te
	}

	//subscribe to MT
	//if we encounter an error we will want to unwind successful subscriptions
	type subscription struct {
		namespace []string
		version   int
	}
	subscriptions := make([]subscription, 0)
	for _, m := range mt {
		ucd, err := metricManager.SubscribeMetric(m.Namespace(), m.Version(), cd)
		if err == nil {
			mtc = append(mtc, &metricType{config: ucd, metricType: m})
			subscriptions = append(subscriptions, subscription{namespace: m.Namespace(), version: m.Version()})
		} else {
			te.errs = append(te.errs, err.Errors()...)
		}
	}

	if len(te.errs) > 0 {
		//unwind successful subscriptions
		for _, sub := range subscriptions {
			metricManager.UnsubscribeMetric(sub.namespace, sub.version)
		}
		return nil, te
	}

	task := NewTask(s, mtc)
	return task, nil
}

func New(m ManagesMetric) *scheduler {
	metricManager = m
	return &scheduler{}
}
