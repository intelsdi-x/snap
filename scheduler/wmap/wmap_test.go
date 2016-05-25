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

package wmap

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWorkflow(t *testing.T) {
	jsonP, _ := ioutil.ReadFile("./sample/1.json")
	yamlP, _ := ioutil.ReadFile("./sample/1.yml")

	Convey("Workflow map", t, func() {

		Convey("from yaml", func() {
			fmt.Println("YAML ---")
			wmap, err := FromYaml(yamlP)
			So(err, ShouldBeNil)
			So(wmap, ShouldNotBeNil)
			wmap, err = FromYaml(1)
			So(err, ShouldNotBeEmpty)
			So(wmap, ShouldBeNil)
			fmt.Println(wmap)
		})

		Convey("from json", func() {
			fmt.Println("JSON ---")
			wmap, err := FromJson(jsonP)
			So(err, ShouldBeNil)
			So(wmap, ShouldNotBeNil)
			wmap, err = FromJson(1)
			So(err, ShouldNotBeEmpty)
			So(wmap, ShouldBeNil)
			fmt.Println(wmap)
		})

		Convey("Sample", func() {
			So(SampleWorkflowMapJson(), ShouldNotBeEmpty)
			So(SampleWorkflowMapYaml(), ShouldNotBeEmpty)
		})

		Convey("from json/CollectNode.GetTags()", func() {
			wmap, _ := FromJson(jsonP)
			tags := wmap.CollectNode.GetTags()
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

		Convey("from yaml/CollectNode.GetTags()", func() {
			wmap, _ := FromYaml(jsonP)
			tags := wmap.CollectNode.GetTags()
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

		Convey("NewWorkFlowMap()/GetRequestedMetrics()", func() {
			wmap := NewWorkflowMap()
			So(wmap, ShouldNotBeNil)
			fmt.Println(wmap)
			fmt.Printf("Metrics: %v", wmap.CollectNode.Metrics)
			So(wmap.CollectNode.GetMetrics(), ShouldBeEmpty)
			wmap.CollectNode.AddMetric("/foo/bar", 1)
			So(wmap.CollectNode.GetMetrics(), ShouldNotBeEmpty)
			wmap.CollectNode.GetMetrics()[0].Namespace()
			So(wmap.CollectNode.GetMetrics()[0].Namespace(), ShouldResemble, []string{"foo", "bar"})
			wmap.CollectNode.GetMetrics()[0].Version()
			So(wmap.CollectNode.GetMetrics()[0].Version(), ShouldResemble, 1)
		})

		Convey("AddMetric()/AddConfigItem()", func() {
			wmap := NewWorkflowMap()
			So(wmap, ShouldNotBeNil)
			fmt.Println(wmap)
			fmt.Printf("Metrics: %v\n", wmap.CollectNode.Metrics)
			fmt.Printf("Config : %v\n", wmap.CollectNode.Config)
			So(wmap.CollectNode.Metrics, ShouldBeEmpty)
			wmap.CollectNode.AddMetric("/foo/bar", 1)
			fmt.Printf("Metrics: %v\n", wmap.CollectNode.Metrics)
			So(wmap.CollectNode.Metrics, ShouldNotBeEmpty)
			So(wmap.CollectNode.Config, ShouldBeEmpty)
			wmap.CollectNode.AddConfigItem("/foo/bar", "user", "bob")
			fmt.Printf("Config : %v\n", wmap.CollectNode.Config)
			So(wmap.CollectNode.Config, ShouldNotBeEmpty)
			fmt.Println(wmap)
		})

		Convey("Add()/New Process/New Publish nodes", func() {
			wmap := NewWorkflowMap()
			wmap.CollectNode.AddConfigItem("/foo/bar", "user", "stu")
			fmt.Println(wmap)
			pr1 := &ProcessWorkflowMapNode{
				Name:    "oslo",
				Version: 1,
				Config:  make(map[string]interface{}),
			}
			pr1.Config["version"] = "kilo"
			//NewProcessNode, NewPublishNode
			pr2 := NewProcessNode("floor", 1)
			pu1 := NewPublishNode("isis", 1)
			pu2 := NewPublishNode("zorro", 1)
			//Collect Node Add
			wmap.CollectNode.Add(pr1)              //case process node
			wmap.CollectNode.Add(pu1)              //case publish node
			wmap.CollectNode.Add(wmap.CollectNode) //case default
			So(wmap.CollectNode.ProcessNodes, ShouldNotBeEmpty)
			So(wmap.CollectNode.PublishNodes, ShouldNotBeEmpty)
			//Process Node Add
			wmap.CollectNode.ProcessNodes[0].Add(pr2)
			wmap.CollectNode.ProcessNodes[0].Add(pu2)
			wmap.CollectNode.ProcessNodes[0].Add(wmap.CollectNode)
			So(wmap.CollectNode.ProcessNodes[0].ProcessNodes, ShouldNotBeEmpty)
			So(wmap.CollectNode.ProcessNodes[0].PublishNodes, ShouldNotBeEmpty)
			fmt.Println(wmap)
			//GetConfigNode() nil case
			cn, err := wmap.CollectNode.ProcessNodes[0].ProcessNodes[0].GetConfigNode()
			So(cn, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			cn, err = wmap.CollectNode.PublishNodes[0].GetConfigNode()
			So(cn, ShouldNotBeEmpty)
			So(err, ShouldBeNil)

		})

		Convey("Gets the config tree and the config node", func() {
			wmap := NewWorkflowMap()
			fmt.Println(wmap)
			wmap.CollectNode.AddConfigItem("/foo/bar", "user", "stu")
			pu1 := NewPublishNode("stuff", 1)
			pr1 := NewProcessNode("name", 1)
			pr2 := NewProcessNode("thing", 1)
			pr3 := NewProcessNode("thing", 1)
			wmap.CollectNode.Add(pu1)
			wmap.CollectNode.Add(pr1)
			wmap.CollectNode.Add(pr2)
			wmap.CollectNode.ProcessNodes[0].Add(pr3)
			wmap.CollectNode.PublishNodes[0].AddConfigItem("key", 1)
			wmap.CollectNode.ProcessNodes[0].AddConfigItem("key", 3.14)
			wmap.CollectNode.ProcessNodes[1].AddConfigItem("key", true)
			wmap.CollectNode.ProcessNodes[0].ProcessNodes[0].AddConfigItem("key", struct{}{})

			pu1conf, err2 := wmap.CollectNode.PublishNodes[0].GetConfigNode()
			So(pu1conf, ShouldNotBeEmpty)
			So(err2, ShouldBeNil)

			pr1conf, err3 := wmap.CollectNode.ProcessNodes[0].GetConfigNode()
			So(pr1conf, ShouldNotBeEmpty)
			So(err3, ShouldBeNil)

			pr2conf, err3 := wmap.CollectNode.ProcessNodes[1].GetConfigNode()
			So(pr2conf, ShouldNotBeEmpty)
			So(err3, ShouldBeNil)

			pr3conf, err4 := wmap.CollectNode.ProcessNodes[0].ProcessNodes[0].GetConfigNode()
			So(pr3conf, ShouldNotBeEmpty)
			So(err4, ShouldNotBeNil)

			ctree, err := wmap.CollectNode.GetConfigTree()
			So(ctree, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
			fmt.Println(wmap)
		})

		Convey("Converts strings to bytes or keeps byte type", func() {
			p, err := inStringBytes("test")
			So(p, ShouldResemble, []byte("test"))
			So(err, ShouldBeNil)
			p, err = inStringBytes(1)
			So(p, ShouldBeEmpty)
			So(err, ShouldNotBeNil)
		})

	})
}
