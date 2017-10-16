// +build legacy

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
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"

	"github.com/intelsdi-x/snap/core/serror"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

type mockCollector struct{}

func (m *mockCollector) CollectMetrics(string, map[string]map[string]string) ([]core.Metric, []error) {
	return nil, nil
}

func (m *mockCollector) ExpandWildcards(core.Namespace) ([]core.Namespace, serror.SnapError) {
	return nil, nil
}

func TestCollectorJob(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	cdt := cdata.NewTree()
	//TODO: kromar do something with tags?
	tags := map[string]map[string]string{}
	Convey("newCollectorJob()", t, func() {
		Convey("it returns an init-ed collectorJob", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			So(cj, ShouldHaveSameTypeAs, &collectorJob{})
		})
	})
	Convey("StartTime()", t, func() {
		Convey("it should return the job starttime", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			So(cj.StartTime(), ShouldHaveSameTypeAs, time.Now())
		})
	})
	Convey("Deadline()", t, func() {
		Convey("it should return the job daedline", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			So(cj.Deadline(), ShouldResemble, cj.(*collectorJob).deadline)
		})
	})
	Convey("Type()", t, func() {
		Convey("it should return the job type", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			So(cj.Type(), ShouldEqual, collectJobType)
		})
	})
	Convey("Errors()", t, func() {
		Convey("it should return the errors from the job", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			So(cj.Errors(), ShouldResemble, []error{})
		})
	})
	Convey("AddErrors()", t, func() {
		Convey("it should append errors to the job", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			So(cj.Errors(), ShouldResemble, []error{})

			e1 := errors.New("1")
			e2 := errors.New("2")
			e3 := errors.New("3")

			cj.AddErrors(e1)
			So(cj.Errors(), ShouldResemble, []error{e1})
			cj.AddErrors(e2, e3)
			So(cj.Errors(), ShouldResemble, []error{e1, e2, e3})
		})
	})
	Convey("Run()", t, func() {
		Convey("it should complete without errors", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			cj.(*collectorJob).Run()
			So(cj.Errors(), ShouldResemble, []error{})
		})
	})
}

func TestQueuedJob(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	cdt := cdata.NewTree()
	tags := map[string]map[string]string{}
	Convey("Job()", t, func() {
		Convey("it should return the underlying job", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			qj := newQueuedJob(cj)
			So(qj.Job(), ShouldEqual, cj)
		})
	})
	Convey("Promise()", t, func() {
		Convey("it should return the underlying promise", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt, "taskid", tags)
			qj := newQueuedJob(cj)
			So(qj.Promise().IsComplete(), ShouldBeFalse)
		})
	})
}
