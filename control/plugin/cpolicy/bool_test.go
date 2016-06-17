// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

func TestConfigPolicyRuleBool(t *testing.T) {
	Convey("NewBoolRule", t, func() {

		Convey("empty key", func() {
			r, e := NewBoolRule("", true)
			So(r, ShouldBeNil)
			So(e, ShouldResemble, EmptyKeyError)
		})

		Convey("key is correct", func() {
			r, e := NewBoolRule("thekey", true)
			So(r.Key(), ShouldEqual, "thekey")
			So(e, ShouldBeNil)
		})

		Convey("required is set", func() {
			r, e := NewBoolRule("thekey", true)
			So(r.Required(), ShouldEqual, true)
			So(e, ShouldBeNil)
		})

		Convey("default is set", func() {
			r, e := NewBoolRule("thekey", true, true)
			So(r.Default(), ShouldNotBeNil)
			So(r.Default().Type(), ShouldEqual, "bool")
			So(r.Default().(ctypes.ConfigValueBool).Value, ShouldEqual, true)
			So(e, ShouldBeNil)
		})

		Convey("default is unset", func() {
			r, e := NewBoolRule("thekey", true)
			So(r.Default(), ShouldBeNil)
			So(e, ShouldBeNil)
		})

		Convey("processing", func() {

			Convey("passes with string config value", func() {
				r, e := NewBoolRule("thekey", true, true)
				So(r.Default().Type(), ShouldEqual, "bool")
				So(r.Default().(ctypes.ConfigValueBool).Value, ShouldEqual, true)
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueBool{Value: true}

				e = r.Validate(v)
				So(e, ShouldBeNil)
			})

			Convey("errors with non-string config value", func() {
				r, e := NewBoolRule("thekey", true, true)
				So(r.Default().Type(), ShouldEqual, "bool")
				So(r.Default().(ctypes.ConfigValueBool).Value, ShouldEqual, true)
				So(e, ShouldBeNil)

				v := ctypes.ConfigValueInt{Value: 1}

				e = r.Validate(v)
				So(e, ShouldResemble, errors.New("type mismatch (thekey wanted type 'bool' but provided type 'integer')"))
			})

		})

	})
}
