// + build small

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

package wmap

import (
	"io/ioutil"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/scheduler/wmap/fixtures"
)

func TestWorkflowFromYAML(t *testing.T) {
	Convey("Workflow map from yaml", t, func() {
		wmap, err := FromYaml(fixtures.TaskYAML)
		So(err, ShouldBeNil)
		So(wmap, ShouldNotBeNil)

		wmap, err = FromYaml(1)
		So(err, ShouldNotBeEmpty)
		So(wmap, ShouldBeNil)
	})
}

func TestWorkflowFromJSON(t *testing.T) {
	Convey("Workflow map from json", t, func() {
		wmap, err := FromJson(fixtures.TaskJSON)
		So(err, ShouldBeNil)
		So(wmap, ShouldNotBeNil)

		wmap, err = FromJson(1)
		So(err, ShouldNotBeEmpty)
		So(wmap, ShouldBeNil)
	})
}

func TestSampleWorkflows(t *testing.T) {
	Convey("Sampling workflow map to json", t, func() {
		So(SampleWorkflowMapJson(), ShouldNotBeEmpty)
	})

	Convey("Sampling workflow map to yaml", t, func() {
		So(SampleWorkflowMapYaml(), ShouldNotBeEmpty)
	})

}

func TestTagsOnWorkflow(t *testing.T) {
	Convey("Extracting tags from workflow", t, func() {
		Convey("From JSON", func() {
			wmap, _ := FromJson(fixtures.TaskJSON)
			tags := wmap.Collect.GetTags()
			So(tags, ShouldNotBeNil)
			So(tags, ShouldResemble, map[string]map[string]string{
				"/foo/bar": {
					"tag1": "val1",
					"tag2": "val2",
				},
				"/foo/baz": {
					"tag3": "val3",
				},
			})
		})

		Convey("From YAML", func() {
			wmap, _ := FromYaml(fixtures.TaskYAML)
			tags := wmap.Collect.GetTags()
			So(tags, ShouldNotBeNil)
			So(tags, ShouldResemble, map[string]map[string]string{
				"/foo/bar": {
					"tag1": "val1",
					"tag2": "val2",
				},
				"/foo/baz": {
					"tag3": "val3",
				},
			})
		})

	})
}

func TestWfGetRequestedMetrics(t *testing.T) {
	Convey("NewWorkFlowMap()/GetRequestedMetrics()", t, func() {
		wmap := NewWorkflowMap()
		So(wmap, ShouldNotBeNil)
		So(wmap.Collect.GetMetrics(), ShouldBeEmpty)
		wmap.Collect.AddMetric("/foo/bar", 1)
		So(wmap.Collect.GetMetrics(), ShouldNotBeEmpty)
		wmap.Collect.GetMetrics()[0].Namespace()
		So(wmap.Collect.GetMetrics()[0].Namespace(), ShouldResemble, []string{"foo", "bar"})
		wmap.Collect.GetMetrics()[0].Version()
		So(wmap.Collect.GetMetrics()[0].Version(), ShouldResemble, 1)
	})
}

func TestWfAddConfigItem(t *testing.T) {
	Convey("AddMetric()/AddConfigItem()", t, func() {
		wmap := NewWorkflowMap()
		So(wmap, ShouldNotBeNil)
		So(wmap.Collect.Metrics, ShouldBeEmpty)
		wmap.Collect.AddMetric("/foo/bar", 1)
		So(wmap.Collect.Metrics, ShouldNotBeEmpty)
		So(wmap.Collect.Config, ShouldBeEmpty)
		wmap.Collect.AddConfigItem("/foo/bar", "user", "bob")
		So(wmap.Collect.Config, ShouldNotBeEmpty)
	})
}

func TestWfPublishProcessNodes(t *testing.T) {
	Convey("Add()/New Process/New Publish nodes", t, func() {
		wmap := NewWorkflowMap()
		wmap.Collect.AddConfigItem("/foo/bar", "user", "stu")

		pr1 := &ProcessWorkflowMapNode{
			PluginName:    "oslo",
			PluginVersion: 1,
			Config:        make(map[string]interface{}),
		}

		pr1.Config["version"] = "kilo"

		//NewProcessNode, NewPublishNode
		pr2 := NewProcessNode("floor", 1)
		pu1 := NewPublishNode("isis", 1)
		pu2 := NewPublishNode("zorro", 1)

		//Collect Node Add
		wmap.Collect.Add(pr1)          //case process node
		wmap.Collect.Add(pu1)          //case publish node
		wmap.Collect.Add(wmap.Collect) //case default

		So(wmap.Collect.Process, ShouldNotBeEmpty)
		So(wmap.Collect.Publish, ShouldNotBeEmpty)

		//Process Node Add
		wmap.Collect.Process[0].Add(pr2)
		wmap.Collect.Process[0].Add(pu2)
		wmap.Collect.Process[0].Add(wmap.Collect)

		So(wmap.Collect.Process[0].Process, ShouldNotBeEmpty)
		So(wmap.Collect.Process[0].Publish, ShouldNotBeEmpty)

		//GetConfigNode() nil case
		cn, err := wmap.Collect.Process[0].Process[0].GetConfigNode()
		So(cn, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
		cn, err = wmap.Collect.Publish[0].GetConfigNode()
		So(cn, ShouldNotBeEmpty)
		So(err, ShouldBeNil)

	})

}

func TestWfGetConfigNodeTree(t *testing.T) {
	Convey("Gets the config tree and the config node", t, func() {
		wmap := NewWorkflowMap()
		wmap.Collect.AddConfigItem("/foo/bar", "user", "stu")
		pu1 := NewPublishNode("stuff", 1)
		pr1 := NewProcessNode("name", 1)
		pr2 := NewProcessNode("thing", 1)
		pr3 := NewProcessNode("thing", 1)

		wmap.Collect.Add(pu1)
		wmap.Collect.Add(pr1)
		wmap.Collect.Add(pr2)

		wmap.Collect.Process[0].Add(pr3)
		wmap.Collect.Publish[0].AddConfigItem("key", 1)
		wmap.Collect.Process[0].AddConfigItem("key", 3.14)
		wmap.Collect.Process[1].AddConfigItem("key", true)
		wmap.Collect.Process[0].Process[0].AddConfigItem("key", struct{}{})

		pu1conf, err2 := wmap.Collect.Publish[0].GetConfigNode()
		So(pu1conf, ShouldNotBeEmpty)
		So(err2, ShouldBeNil)

		pr1conf, err3 := wmap.Collect.Process[0].GetConfigNode()
		So(pr1conf, ShouldNotBeEmpty)
		So(err3, ShouldBeNil)

		pr2conf, err3 := wmap.Collect.Process[1].GetConfigNode()
		So(pr2conf, ShouldNotBeEmpty)
		So(err3, ShouldBeNil)

		pr3conf, err4 := wmap.Collect.Process[0].Process[0].GetConfigNode()
		So(pr3conf, ShouldNotBeEmpty)
		So(err4, ShouldNotBeNil)

		ctree, err := wmap.Collect.GetConfigTree()
		So(ctree, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestStringByteConvertion(t *testing.T) {
	Convey("Converts strings to bytes or keeps byte type", t, func() {
		p, err := inStringBytes("test")
		So(p, ShouldResemble, []byte("test"))
		So(err, ShouldBeNil)
		p, err = inStringBytes(1)
		So(p, ShouldBeEmpty)
		So(err, ShouldNotBeNil)
	})
}

func TestMetricSeparator(t *testing.T) {
	jsonP, _ := ioutil.ReadFile("./sample/2.json")

	Convey("Get Metric", t, func() {
		Convey("from json", func() {
			wmap, err := FromJson(jsonP)
			So(err, ShouldBeNil)
			So(wmap, ShouldNotBeNil)

			mts := wmap.Collect.GetMetrics()
			for i, m := range mts {
				Convey("namespace "+strconv.Itoa(i), func() {
					So(len(m.Namespace()), ShouldEqual, 2)
				})
			}
		})
	})
}
