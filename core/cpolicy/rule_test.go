package cpolicy

import (
	"testing"

	"github.com/intelsdilabs/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicyRule(t *testing.T) {
	Convey("NewStringRule", t, func() {

		Convey("empty key", func() {
			r, e := NewStringRule("", true)
			So(r, ShouldBeNil)
			So(e, ShouldResemble, EmptyKeyError)
		})

		Convey("key is correct", func() {
			r, e := NewStringRule("thekey", true)
			So(r.Key(), ShouldEqual, "thekey")
			So(e, ShouldBeNil)
		})

		Convey("required is set", func() {
			r, e := NewStringRule("thekey", true)
			So(r.(*stringRule).required, ShouldEqual, true)
			So(e, ShouldBeNil)
		})

		Convey("default exists and points to string", func() {
			r, e := NewStringRule("thekey", true, "wat")
			So(*(r.(*stringRule).default_), ShouldEqual, "wat")
			So(e, ShouldBeNil)
		})

		Convey("processing", func() {

			Convey("passes with string config value", func() {
				r, e := NewStringRule("thekey", true, "wat")
				So(*(r.(*stringRule).default_), ShouldEqual, "wat")
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueStr{Value: "foo"}

				e = r.Validate(v)
				So(e, ShouldBeNil)
			})

			Convey("errors with non-string config value", func() {
				r, e := NewStringRule("thekey", true, "wat")
				So(*(r.(*stringRule).default_), ShouldEqual, "wat")
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueInt{Value: 1}

				e = r.Validate(v)
				So(e, ShouldResemble, WrongTypeError)
			})

		})

	})
}
