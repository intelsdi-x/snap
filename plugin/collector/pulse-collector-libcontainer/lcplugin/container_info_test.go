package lcplugin

import (
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetConfig(t *testing.T) {

	Convey("TestGetConfig with container.json fixture", t, func() {

		pwd, err := os.Getwd() // maybe I should use PULSE_PATH?
		So(err, ShouldBeNil)

		containerJsonPth := path.Join(pwd, "test_fixtures", "container.json")

		config, err := getConfig(containerJsonPth)
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		So(config.ProcessLabel, ShouldEqual, "system_u:system_r:svirt_lxc_net_t:s0:c264,c636")
		So(config.Hostname, ShouldEqual, "014fe3fbf5e6")
	})
}

func TestGetState(t *testing.T) {

	Convey("TestGetState with state.json fixture", t, func() {

		pwd, err := os.Getwd() // maybe I should use PULSE_PATH?
		So(err, ShouldBeNil)

		stateJsonPth := path.Join(pwd, "test_fixtures", "state.json")

		state, err := getState(stateJsonPth)
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		So(state.InitPid, ShouldEqual, 8059)
		So(state.InitStartTime, ShouldEqual, "955735")
	})
}
