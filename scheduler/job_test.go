/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

type mockCollector struct{}

func (m *mockCollector) CollectMetrics([]core.Metric, time.Time) ([]core.Metric, []error) {
	return nil, nil
}

func TestCollectorJob(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	cdt := cdata.NewTree()
	Convey("newCollectorJob()", t, func() {
		Convey("it returns an init-ed collectorJob", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			So(cj, ShouldHaveSameTypeAs, &collectorJob{})
		})
	})
	Convey("StartTime()", t, func() {
		Convey("it should return the job starttime", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			So(cj.StartTime(), ShouldHaveSameTypeAs, time.Now())
		})
	})
	Convey("Deadline()", t, func() {
		Convey("it should return the job daedline", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			So(cj.Deadline(), ShouldResemble, cj.(*collectorJob).deadline)
		})
	})
	Convey("Type()", t, func() {
		Convey("it should return the job type", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			So(cj.Type(), ShouldEqual, collectJobType)
		})
	})
	Convey("ReplChan()", t, func() {
		Convey("it should return the reply channel", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			So(cj.ReplChan(), ShouldHaveSameTypeAs, make(chan struct{}))
		})
	})
	// Convey("Metrics()", t, func() {
	// 	Convey("it should return the job metrics", func() {
	// 		cj := newCollectorJob([]core.MetricType{}, defaultDeadline, &mockCollector{})
	// 		So(cj.Metrics(), ShouldResemble, []core.Metric{})
	// 	})
	// })
	Convey("Errors()", t, func() {
		Convey("it should return the errors from the job", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			So(cj.Errors(), ShouldResemble, []error{})
		})
	})
	Convey("Run()", t, func() {
		Convey("it should reply on the reply chan", func() {
			cj := newCollectorJob([]core.RequestedMetric{}, defaultDeadline, &mockCollector{}, cdt)
			go cj.(*collectorJob).Run()
			<-cj.(*collectorJob).replchan
			So(cj.Errors(), ShouldResemble, []error{})
		})
	})
}
