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
	"testing"

	"github.com/intelsdi-x/snap/scheduler/wmap"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflowString(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	Convey("String", t, func() {
		w := wmap.NewWorkflowMap()
		w.Collect.AddMetric("fall", 1)
		pr1 := wmap.NewProcessNode("winter", 1)
		pr2 := wmap.NewProcessNode("summer", 1)
		pu1 := wmap.NewPublishNode("spring", 1)
		pu2 := wmap.NewPublishNode("autumn", 1)
		w.Collect.AddConfigItem("/foo/bar", "user", "rain")
		pr1.AddConfigItem("leaves", 1)
		pr2.AddConfigItem("flowers", 2)
		pu2.AddConfigItem("grass", 3)
		w.Collect.Add(pr1)
		w.Collect.Add(pu1)
		w.Collect.Process[0].Add(pr2)
		w.Collect.Process[0].Add(pu2)

		wf, err := wmapToWorkflow(w)
		So(err, ShouldBeNil)
		str := wf.String()
		//fmt.Printf("%v", str)
		So(str, ShouldNotBeEmpty)
		str2 := wf.processNodes[0].String("")
		So(str2, ShouldNotBeEmpty)
		metricString("", wf.metrics)
	})
}
