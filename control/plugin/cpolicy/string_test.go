package cpolicy

import (
	"errors"
	"testing"

	"github.com/intelsdilabs/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicyRuleString(t *testing.T) {
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
			So(r.Required(), ShouldEqual, true)
			So(e, ShouldBeNil)
		})

		Convey("default is set", func() {
			r, e := NewStringRule("thekey", true, "wat")
			So(r.Default(), ShouldNotBeNil)
			So(r.Default().Type(), ShouldEqual, "string")
			So(r.Default().(*ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
			So(e, ShouldBeNil)
		})

		Convey("default is unset", func() {
			r, e := NewStringRule("thekey", true)
			So(r.Default(), ShouldBeNil)
			So(e, ShouldBeNil)
		})

		Convey("processing", func() {

			Convey("passes with string config value", func() {
				r, e := NewStringRule("thekey", true, "wat")
				So(r.Default().Type(), ShouldEqual, "string")
				So(r.Default().(*ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueStr{Value: "foo"}

				e = r.Validate(v)
				So(e, ShouldBeNil)
			})

			Convey("errors with non-string config value", func() {
				r, e := NewStringRule("thekey", true, "wat")
				So(r.Default().Type(), ShouldEqual, "string")
				So(r.Default().(*ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueInt{Value: 1}

				e = r.Validate(v)
				So(e, ShouldResemble, errors.New("type mismatch (thekey wanted type 'string' but provided type 'integer')"))
			})

		})

	})
}
