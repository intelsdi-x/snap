package availability

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// Uses the dummy collector plugin to simulate Loading
func TestRunner(t *testing.T) {
	Convey("pulse/control/availability", t, func() {

		Convey("Runner", func() {

			Convey(".AddEmitter", func() {

				Convey("Create a Runner and add emitter to it", func() {})

				Convey("Create a Runner and add multiple emitters to it", func() {})

			})

			Convey(".Start", func() {

				Convey("Starting Runner without adding emitters", func() {})

				Convey("Starting Runner after adding one emitter", func() {})

				Convey("Starting Runner after adding multiple emitter", func() {})

			})

			Convey(".Stop", func() {

				Convey("Stopping runner removes handlers from emitters", func() {})

				Convey("Stopping runner stops all AvailablePlugins running", func() {})

			})

		})
	})
}
