package cpolicy

import (
	"errors"
	"testing"

	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicyRuleFloat(t *testing.T) {
	Convey("NewFloatRule", t, func() {

		Convey("empty key", func() {
			r, e := NewFloatRule("", true)
			So(r, ShouldBeNil)
			So(e, ShouldResemble, EmptyKeyError)
		})

		Convey("key is correct", func() {
			r, e := NewFloatRule("thekey", true)
			So(r.Key(), ShouldEqual, "thekey")
			So(e, ShouldBeNil)
		})

		Convey("required is set", func() {
			r, e := NewFloatRule("thekey", true)
			So(r.Required(), ShouldEqual, true)
			So(e, ShouldBeNil)
		})

		Convey("default is set", func() {
			r, e := NewFloatRule("thekey", true, 7)
			So(r.Default(), ShouldNotBeNil)
			So(r.Default().Type(), ShouldEqual, "float")
			So(r.Default().(*ctypes.ConfigValueFloat).Value, ShouldEqual, 7)
			So(e, ShouldBeNil)
		})

		Convey("default is unset", func() {
			r, e := NewFloatRule("thekey", true)
			So(r.Default(), ShouldBeNil)
			So(e, ShouldBeNil)
		})

		Convey("min is set", func() {
			r, e := NewFloatRule("thekey", true)
			r.SetMinimum(0)
			So(*r.minimum, ShouldEqual, 0)
			So(e, ShouldBeNil)
		})

		Convey("max is set", func() {
			r, e := NewFloatRule("thekey", true, 1)
			r.SetMaximum(127)
			So(*r.maximum, ShouldEqual, 127)
			So(e, ShouldBeNil)
		})

		Convey("processing", func() {

			Convey("passes with float config value", func() {
				r, e := NewFloatRule("thekey", true, 7)
				So(r.Default(), ShouldNotBeNil)
				So(r.Default().Type(), ShouldEqual, "float")
				So(r.Default().(*ctypes.ConfigValueFloat).Value, ShouldEqual, 7)
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueFloat{Value: 1}

				e = r.Validate(v)
				So(e, ShouldBeNil)
			})

			Convey("error with non-float config value", func() {
				r, e := NewFloatRule("thekey", true, 2)
				So(r.Default(), ShouldNotBeNil)
				So(r.Default().Type(), ShouldEqual, "float")
				So(r.Default().(*ctypes.ConfigValueFloat).Value, ShouldEqual, 2)
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueStr{Value: "wat"}

				e = r.Validate(v)
				So(e, ShouldResemble, errors.New("type mismatch (thekey wanted type 'float' but provided type 'string')"))
			})

			Convey("error with value below minimum", func() {
				r, e := NewFloatRule("thekey", true)
				r.SetMinimum(1.987)
				So(*r.minimum, ShouldEqual, 1.987)
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueFloat{Value: 0.23}

				e = r.Validate(v)
				So(e, ShouldResemble, errors.New("value is under minimum (thekey value 0.230000 < 1.987000)"))
			})

			Convey("error with value above maximum", func() {
				r, e := NewFloatRule("thekey", true)
				r.SetMaximum(127.127)
				So(*r.maximum, ShouldEqual, 127.127)
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueFloat{Value: 200.000001}

				e = r.Validate(v)
				So(e, ShouldResemble, errors.New("value is over maximum (thekey value 200.000001 > 127.127000)"))
			})

		})

	})
}
