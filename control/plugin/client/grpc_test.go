// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package client

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestToCoreMetric(t *testing.T) {
	tc := testCases()
	Convey("Test ToCoreMetric", t, func() {
		for _, c := range tc {
			Convey(c.description, func() {
				mt := ToMetric(c)
				cmt := ToCoreMetric(mt)
				So(cmt.Timestamp(), ShouldNotBeNil)
				So(cmt.LastAdvertisedTime(), ShouldNotBeNil)

				if cmt.Version() == 2 {
					So(cmt.Timestamp(), ShouldResemble, cmt.LastAdvertisedTime())
				}
			})
		}
	})
}

func testCases() []*metric {
	now := time.Now()
	tc := []*metric{
		&metric{
			namespace:   core.NewNamespace("a", "b", "c"),
			version:     1,
			description: "No timeStamp and lastAdvertisedTime defined",
		},
		&metric{
			namespace:   core.NewNamespace("x", "y", "z"),
			version:     1,
			timeStamp:   time.Now(),
			description: "Has timestamp but no lastAdvertisedTime defined",
		},
		&metric{
			namespace:          core.NewNamespace("x", "y", "z"),
			version:            1,
			lastAdvertisedTime: time.Now(),
			description:        "No timestamp but has lastAdvertisedTime defined",
		},
		&metric{
			namespace:          core.NewNamespace("x", "y", "z"),
			version:            2,
			timeStamp:          now,
			lastAdvertisedTime: now,
			description:        "Has both timestamp and lastAdvertisedTime defined",
		},
	}
	return tc
}
