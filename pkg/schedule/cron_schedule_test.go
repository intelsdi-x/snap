// +build legacy

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

package schedule

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCronSchedule(t *testing.T) {
	Convey("Cron Schedule", t, func() {
		Convey("valid cron entry", func() {
			i := "0 * * * * *"
			c := NewCronSchedule(i)
			e := c.Validate()
			So(e, ShouldBeNil)
		})
		Convey("missing cron entry", func() {
			i := ""
			c := NewCronSchedule(i)
			e := c.Validate()
			So(e, ShouldEqual, ErrMissingCronEntry)
		})
		Convey("invalid cron entry", func() {
			i := "invalid cron entry"
			c := NewCronSchedule(i)
			e := c.Validate()
			So(e, ShouldNotBeNil)
		})
		Convey("wait on valid cron entry", func() {
			i := "@every 1s"
			c := NewCronSchedule(i)
			now := time.Now()
			r := c.Wait(now)
			So(r, ShouldNotBeNil)
			So(r.State(), ShouldEqual, Active)
			So(r.Error(), ShouldBeNil)
			So(r.Missed(), ShouldEqual, 0)
			lastTime := r.LastTime()
			l := lastTime.After(now) && lastTime.Before(time.Now())
			So(l, ShouldBeTrue)
		})
		Convey("counting misses in Wait()", func() {
			i := "@every 1s"
			c := NewCronSchedule(i)
			now := time.Now()
			r := c.Wait(now)
			then := now.Add(-time.Duration(10) * time.Second)
			r = c.Wait(then)
			So(r, ShouldNotBeNil)
			So(r.State(), ShouldEqual, Active)
			So(r.Error(), ShouldBeNil)
			So(r.Missed(), ShouldBeBetweenOrEqual, 10, 12)
			lastTime := r.LastTime()
			l := lastTime.After(now) && lastTime.Before(time.Now())
			So(l, ShouldBeTrue)
		})
	})
}
