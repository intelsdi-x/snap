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

package cpolicy

import (
	"errors"
	"testing"

	"github.com/intelsdi-x/snap/core/ctypes"
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
			So(r.Default().(ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
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
				So(r.Default().(ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueStr{Value: "foo"}

				e = r.Validate(v)
				So(e, ShouldBeNil)
			})

			Convey("errors with non-string config value", func() {
				r, e := NewStringRule("thekey", true, "wat")
				So(r.Default().Type(), ShouldEqual, "string")
				So(r.Default().(ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueInt{Value: 1}

				e = r.Validate(v)
				So(e, ShouldResemble, errors.New("type mismatch (thekey wanted type 'string' but provided type 'integer')"))
			})

		})

	})
}
