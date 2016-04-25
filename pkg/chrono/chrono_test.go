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

package chrono

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestChrono(t *testing.T) {
	Convey("Given a new chrono", t, func() {
		var Chrono chrono
		Convey("Forward time should be 0", func() {
			So(Chrono.skew, ShouldEqual, 0)
		})

		Convey("When forwarded for 30 seconds", func() {
			Chrono.Forward(30 * time.Second)
			Convey("Forward time should be 30 seconds", func() {
				So(Chrono.skew, ShouldEqual, 30*time.Second)
			})
		})

		Convey("After pausing time", func() {
			Chrono.Pause()
			before := Chrono.Now()
			Convey("And waiting 10 milliseconds", func() {
				time.Sleep(10 * time.Millisecond)

				Convey("Time should stand still", func() {
					So(Chrono.Now().Equal(before), ShouldBeTrue)
				})
			})

			Convey("When forwarding another hour", func() {
				Chrono.Forward(1 * time.Hour)

				Convey("Time should be exactly when we stopped plus one hour", func() {
					So(Chrono.Now().Equal(before.Add(1*time.Hour)), ShouldBeTrue)
				})
			})

			Convey("When resetting time", func() {
				Chrono.Reset()
				Convey("Forward time should be 0", func() {
					So(Chrono.skew, ShouldEqual, 0)
				})

				Convey("And time should be the same as when we paused", func() {
					So(Chrono.Now().Equal(before), ShouldBeTrue)
				})
			})
		})

		Convey("Continuing time", func() {
			Chrono.Continue()
			before := Chrono.Now()
			Convey("And waiting 10 milliseconds", func() {
				time.Sleep(10 * time.Millisecond)

				Convey("Time should progress", func() {
					So(Chrono.Now().After(before), ShouldBeTrue)
				})
			})
		})
	})
}
