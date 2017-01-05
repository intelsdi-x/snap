// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015-2016 Intel Corporation

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

package control

import (
	"testing"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	foo = "/intel/foo"
	bar = "/intel/foo/bar"
	tar = "/intel/tar/qaz"
)

func TestAddTagsFromWorkflow(t *testing.T) {
	hostnameReader = &mockHostnameReader{}
	tcs := prepareTestCases()
	Convey("Adding tags to metric type", t, func() {
		p := newPluginManager()
		for _, tc := range tcs {
			outputTags := p.AddStandardAndWorkflowTags(tc.Metric, tc.InputTags).Tags()
			So(outputTags, ShouldNotBeNil)
			So(outputTags, ShouldResemble, tc.ExpectedTags)
		}
	})
}

func TestContainsTuplePositive(t *testing.T) {
	Convey("when tuple contains two items", t, func() {
		dut := "(host0;host1)"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeTrue)
		So(len(tuple), ShouldEqual, 2)
		So(tuple, ShouldContain, "host0")
		So(tuple, ShouldContain, "host1")
	})
	Convey("when tuple contains three items", t, func() {
		dut := "(host0;host1;host2)"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeTrue)
		So(len(tuple), ShouldEqual, 3)
		So(tuple, ShouldContain, "host0")
		So(tuple, ShouldContain, "host2")
		So(tuple, ShouldContain, "host2")

	})
	Convey("when tuple contains white spaces", t, func() {
		dut := "( host0 ; host1 ; host2 )"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeTrue)
		So(len(tuple), ShouldEqual, 3)
		So(tuple, ShouldContain, "host0")
		So(tuple, ShouldContain, "host2")
		So(tuple, ShouldContain, "host2")

	})
	Convey("when tuple contains multiple-word items", t, func() {
		dut := "(mock host0;the mock host1;this is mock host2)"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeTrue)
		So(len(tuple), ShouldEqual, 3)
		So(tuple, ShouldContain, "mock host0")
		So(tuple, ShouldContain, "the mock host1")
		So(tuple, ShouldContain, "this is mock host2")

	})
	Convey("when tuple's item contains brackets by itself", t, func() {
		dut := "(host0(some details);host1(some details))"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeTrue)
		So(len(tuple), ShouldEqual, 2)
		So(tuple, ShouldContain, "host0(some details)")
		So(tuple, ShouldContain, "host1(some details)")
	})
}

func TestContainsTupleNegative(t *testing.T) {
	Convey("for one-word namespace's element", t, func() {
		dut := "host0"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeFalse)
		So(tuple, ShouldBeNil)
	})
	Convey("for two-word namespace's element", t, func() {
		dut := "my host0"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeFalse)
		So(tuple, ShouldBeNil)
	})
	Convey("missing brackets in tuple", t, func() {
		// that also means that using semicolon in metric namespace is allowed
		// because without brackets it won't be recognized as a tuple
		dut := "host0;host1"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeFalse)
		So(tuple, ShouldBeNil)
	})
	Convey("missing semicolon in tuple", t, func() {
		// that also means that using brackets in metric namespace is allowed
		// because without semicolon used as separator it won't be recognized as a tuple
		dut := "(host0, host1)"
		hasTuple, tuple := containsTuple(dut)
		So(hasTuple, ShouldBeFalse)
		So(tuple, ShouldBeNil)
	})
}

type mockHostnameReader struct{}

func (m *mockHostnameReader) Hostname() string {
	return "hostname"
}

type testCase struct {
	Metric       plugin.MetricType
	InputTags    map[string]map[string]string
	ExpectedTags map[string]string
}

func prepareTestCases() []testCase {
	hostname := hostnameReader.Hostname()
	fooTags := map[string]string{
		"foo_tag": "foo_val",
	}
	barTags := map[string]string{
		"foobar_tag": "foobar_val",
	}
	tarTags := map[string]string{
		"tarqaz_tag": "tarqaz_val",
	}

	allTags := map[string]map[string]string{
		foo: fooTags,
		bar: barTags,
		tar: tarTags,
	}

	foobazMetric := plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "foo", "baz"),
	}
	foobarMetric := plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "foo", "bar"),
	}
	tarqazMetric := plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "tar", "qaz"),
	}

	stdMetric := plugin.MetricType{
		Namespace_: core.NewNamespace("intel", "std"),
	}

	foobazExpected := map[string]string{
		core.STD_TAG_PLUGIN_RUNNING_ON: hostname,
		"foo_tag":                      "foo_val",
	}
	foobarExpected := map[string]string{
		core.STD_TAG_PLUGIN_RUNNING_ON: hostname,
		"foo_tag":                      "foo_val",
		"foobar_tag":                   "foobar_val",
	}
	tarqazExpected := map[string]string{
		core.STD_TAG_PLUGIN_RUNNING_ON: hostname,
		"tarqaz_tag":                   "tarqaz_val",
	}
	stdExpected := map[string]string{
		core.STD_TAG_PLUGIN_RUNNING_ON: hostname,
	}

	testCases := []testCase{
		{foobazMetric, allTags, foobazExpected},
		{foobarMetric, allTags, foobarExpected},
		{tarqazMetric, allTags, tarqazExpected},
		{stdMetric, allTags, stdExpected},
		{foobazMetric, nil, stdExpected},
	}

	return testCases
}
