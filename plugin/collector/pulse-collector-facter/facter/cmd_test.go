// +build linux

// Tests for communication with external cmd facter (executable)

package facter

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// 4 seconds because default time for goconvey (5 seconds for test)
const testFacterTimeout = 4 * time.Second

func TestDefaultConfig(t *testing.T) {
	Convey("check default config", t, func() {
		cmdConfig := newDefaultCmdConfig()
		So(cmdConfig.executable, ShouldEqual, "facter")
		So(cmdConfig.options, ShouldResemble, []string{"--json"})
	})
}

func TestCmdCommunication(t *testing.T) {
	Convey("error when facter binary isn't found", t, func() {
		_, err := getFacts([]string{"whatever"}, testFacterTimeout, &cmdConfig{executable: "wrongbin"})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "file not found") // isn't ubuntu specific ?
	})

	Convey("error when facter output isn't parsable", t, func() {
		_, err := getFacts([]string{"whatever"}, testFacterTimeout, &cmdConfig{executable: "facter", options: []string{}})
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "unexpected end of JSON input")
	})
}

func TestGetFacts(t *testing.T) {

	Convey("getFacts from real facter", t, func() {

		Convey("time outs", func() {
			_, err := getFacts([]string{}, 0*time.Second, nil)
			So(err, ShouldNotBeNil)
		})

		Convey("returns all something within given time", func() {
			facts, err := getFacts([]string{}, testFacterTimeout, nil)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
		})

		Convey("returns right thing when asked eg. kernel => linux", func() {
			// 4 seconds because default time for goconvey
			facts, err := getFacts([]string{"kernel"}, testFacterTimeout, nil)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(len(facts), ShouldEqual, 1)
			fact, exist := facts["kernel"]
			So(exist, ShouldEqual, true)
			So(fact, ShouldNotBeNil)
		})

		Convey("returns nil in fact value when for non existing fact", func() {
			// 4 seconds because default time for goconvey
			facts, err := getFacts([]string{"thereisnosuchfact"}, testFacterTimeout, nil)
			So(err, ShouldBeNil)
			So(facts, ShouldNotBeEmpty)
			So(len(facts), ShouldEqual, 1)
			fact, exist := facts["thereisnosuchfact"]
			So(exist, ShouldEqual, true)
			So(fact, ShouldBeNil)
		})

	})
}
