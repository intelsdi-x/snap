// Tests for communication with external cmd facter

package facter

import (
	"reflect"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetFacts(t *testing.T) {
	Convey("getFacts ", t, func() {

		Convey("time outs", func() {
			_, _, err := getFacts([]string{}, 0*time.Second)
			So(err, ShouldNotBeNil)
		})

		Convey("returns all something within given time", func() {
			start := time.Now()
			// 4 seconds because default time for goconvey
			facts, when, err := getFacts([]string{}, 4*time.Second)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(*when, ShouldHappenBetween, start, time.Now())
		})

		Convey("returns right thing when asked eg. kernel => linux", func() {
			// 4 seconds because default time for goconvey
			facts, _, err := getFacts([]string{"kernel"}, 4*time.Second)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(len(*facts), ShouldEqual, 1)
			fact, exist := (*facts)["kernel"]
			So(exist, ShouldEqual, true)
			So(fact, ShouldNotBeNil)
		})

		Convey("returns nil in fact value when for non existing fact", func() {
			// 4 seconds because default time for goconvey
			facts, _, err := getFacts([]string{"thereisnosuchfact"}, 4*time.Second)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(len(*facts), ShouldEqual, 1)
			fact, exist := (*facts)["thereisnosuchfact"]
			So(exist, ShouldEqual, true)
			So(fact, ShouldBeNil)
		})

	})
}
