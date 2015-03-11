// Tests for communication with external cmd facter (executable)

package facter

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaultConfig(t *testing.T) {
	Convey("check default config", t, func() {
		cmdConfig := newDefaultCmdConfig()
		So(cmdConfig.executable, ShouldEqual, "facter")
		So(cmdConfig.options, ShouldResemble, []string{"--json"})
	})
}

func TestCmdCommunication(t *testing.T) {
	Convey("error when facter binary isn't found", t, func() {
		_, _, err := getFacts([]string{"whatever"}, defaultFacterDeadline, &cmdConfig{executable: "wrongbin"})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "file not found") // isn't ubuntu specific ?
	})

	Convey("error when facter output isn't parsable", t, func() {
		_, _, err := getFacts([]string{"whatever"}, defaultFacterDeadline, &cmdConfig{executable: "facter", options: []string{}})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unexpected end of JSON input")
	})
}

func TestGetFacts(t *testing.T) {
	Convey("getFacts from real facter", t, func() {

		Convey("time outs", func() {
			_, _, err := getFacts([]string{}, 0*time.Second, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("returns all something within given time", func() {
			start := time.Now()
			// 4 seconds because default time for goconvey
			facts, when, err := getFacts([]string{}, 4*time.Second, nil)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(*when, ShouldHappenBetween, start, time.Now())
		})

		Convey("returns right thing when asked eg. kernel => linux", func() {
			// 4 seconds because default time for goconvey
			facts, _, err := getFacts([]string{"kernel"}, 4*time.Second, nil)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(len(*facts), ShouldEqual, 1)
			fact, exist := (*facts)["kernel"]
			So(exist, ShouldEqual, true)
			So(fact, ShouldNotBeNil)
		})

		Convey("returns nil in fact value when for non existing fact", func() {
			// 4 seconds because default time for goconvey
			facts, _, err := getFacts([]string{"thereisnosuchfact"}, 4*time.Second, nil)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(len(*facts), ShouldEqual, 1)
			fact, exist := (*facts)["thereisnosuchfact"]
			So(exist, ShouldEqual, true)
			So(fact, ShouldBeNil)
		})

	})

	// TODO: prepare fake facter executable to better stress our part
	// Convey("getFacts with fake facter")
}
