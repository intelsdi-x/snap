package control

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRouter(t *testing.T) {
	Convey("given a new router created by pluginControl", t, func() {
		// Create controller
		c := New()
		router := c.pluginRouter

		Convey("is not nil", func() {
			So(router, ShouldNotBeNil)
		})

		Convey("has default strategy", func() {
			So(router.Strategy(), ShouldNotBeNil)
		})

	})
}
