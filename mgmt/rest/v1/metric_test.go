// +build small

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

package v1

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseNamespace(t *testing.T) {
	tcs := getNsTestCases()

	Convey("Test parseNamespace", t, func() {
		for _, c := range tcs {
			Convey("Test parseNamespace "+c.input, func() {
				So(c.output, ShouldResemble, parseNamespace(c.input))
			})
		}
	})
}

type nsTestCase struct {
	input  string
	output []string
}

func getNsTestCases() []nsTestCase {
	tcs := []nsTestCase{
		{
			input:  "小a小b小c",
			output: []string{"a", "b", "c"}},
		{
			input:  "%a%b%c",
			output: []string{"a", "b", "c"}},
		{
			input:  "-aヒ-b/-c|",
			output: []string{"aヒ", "b/", "c|"}},
		{
			input:  ">a>b=>c=",
			output: []string{"a", "b=", "c="}},
		{
			input:  ">a>b<>c<",
			output: []string{"a", "b<", "c<"}},
		{
			input:  "㊽a㊽b%㊽c/|",
			output: []string{"a", "b%", "c/|"}},
	}
	return tcs
}
