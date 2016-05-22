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

package strategy

import (
	"testing"
	"time"

	. "github.com/intelsdi-x/snap/control/strategy/fixtures"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStickyRouter(t *testing.T) {
	Convey("Given a sticky router", t, func() {
		router := NewSticky(100 * time.Millisecond)
		So(router, ShouldNotBeNil)
		So(router.String(), ShouldResemble, "sticky")
		Convey("Select a plugin when they are available", func() {
			p1 := NewMockAvailablePlugin().WithName("p1")
			p2 := NewMockAvailablePlugin().WithName("p2")
			// select a plugin, for task1,  given a task and two available plugins
			sp1, err := router.Select([]AvailablePlugin{p1, p2}, "task1")
			So(err, ShouldBeNil)
			So(sp1, ShouldNotBeNil)
			So(sp1, ShouldEqual, p1)
			// change the order of the plugins provided to the select
			sp2, err := router.Select([]AvailablePlugin{p2, p1}, "task1")
			So(err, ShouldBeNil)
			So(sp1, ShouldNotBeNil)
			So(sp2, ShouldEqual, p1)
			// select the other (last) available plugin for task2
			sp3, err := router.Select([]AvailablePlugin{p2, p1}, "task2")
			So(err, ShouldBeNil)
			So(sp3, ShouldNotBeNil)
			So(sp3, ShouldEqual, p2)
			Convey("Select a plugin when there are NONE available", func() {
				plugins := []AvailablePlugin{p1, p2}
				sp, err := router.Select(plugins, "task3")
				So(sp, ShouldBeNil)
				So(err, ShouldEqual, ErrCouldNotSelect)
			})
		})

	})
}
