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
		for _, tc := range tcs {
			outputTags := addStandardAndWorkflowTags(tc.Metric, tc.InputTags).Tags()
			So(outputTags, ShouldNotBeNil)
			So(outputTags, ShouldResemble, tc.ExpectedTags)
		}
	})
}

type mockHostnameReader struct{}

func (m *mockHostnameReader) Hostname() (string, error) {
	return "hostname", nil
}

type testCase struct {
	Metric       plugin.MetricType
	InputTags    map[string]map[string]string
	ExpectedTags map[string]string
}

func prepareTestCases() []testCase {
	hostname, _ := hostnameReader.Hostname()
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
