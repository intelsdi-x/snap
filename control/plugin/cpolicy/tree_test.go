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
	"encoding/gob"
	"testing"

	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicy(t *testing.T) {
	Convey("ConfigPolicy", t, func() {
		cp := New()

		Convey("new config policy", func() {
			So(t, ShouldNotBeNil)
		})

		Convey("stores a policy node", func() {
			cpn := NewPolicyNode()
			r1, _ := NewStringRule("username", false, "root")
			r2, _ := NewStringRule("password", true)
			cpn.Add(r1, r2)
			ns := []string{"one", "two", "potato"}

			cp.Add(ns, cpn)
			Convey("retrieves store policy", func() {
				gc := cp.Get(ns)
				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["username"].Default().(ctypes.ConfigValueStr).Value, ShouldEqual, "root")
				So(gc.rules["password"].Required(), ShouldEqual, true)
			})
			Convey("retrieves all policies", func() {
				policies := cp.GetAll()
				So(len(policies), ShouldEqual, 1)
			})
			Convey("encode & decode", func() {
				gob.Register(NewPolicyNode())
				gob.Register(&StringRule{})
				buf, err := cp.GobEncode()
				So(err, ShouldBeNil)
				So(buf, ShouldNotBeNil)
				cp2 := &ConfigPolicy{}
				err = cp2.GobDecode(buf)
				So(err, ShouldBeNil)
				So(cp2.config, ShouldNotBeNil)
				gc := cp2.Get([]string{"one", "two", "potato"})
				So(gc, ShouldNotBeNil)
				So(gc.rules["username"], ShouldNotBeNil)
				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"].Required(), ShouldEqual, true)
				So(gc.rules["username"].Default(), ShouldNotBeNil)
				So(gc.rules["password"].Default(), ShouldBeNil)
				So(gc.rules["username"].Default().(ctypes.ConfigValueStr).Value, ShouldEqual, "root")
			})

		})

		Convey("stores multiple a policy nodes", func() {
			cpn1 := NewPolicyNode()
			r11, _ := NewStringRule("password", true)
			r12, _ := NewIntegerRule("port", true)
			r13, _ := NewBoolRule("nova", true)
			cpn1.Add(r11, r12, r13)
			ns1 := []string{"one", "two", "potato"}

			cpn2 := NewPolicyNode()
			r21, _ := NewStringRule("password", true)
			r22, _ := NewFloatRule("rate", true)
			r23, _ := NewBoolRule("nova", true)
			cpn2.Add(r21, r22, r23)
			ns2 := []string{"one", "two", "grapefruit"}

			cpn3 := NewPolicyNode()
			r31, _ := NewStringRule("username", false, "root")
			cpn3.Add(r31)
			ns3 := []string{"one", "two"}

			cp.Add(ns1, cpn1)
			cp.Add(ns2, cpn2)
			cp.Add(ns3, cpn3)

			Convey("base node is nil", func() {
				gc := cp.Get([]string{"one"})
				So(gc, ShouldResemble, NewPolicyNode())
			})

			Convey("retrieves all policies", func() {
				policies := cp.GetAll()
				So(len(policies), ShouldEqual, 3)
			})

			Convey("two is correct", func() {
				gc := cp.Get([]string{"one", "two"})
				So(gc, ShouldNotBeNil)

				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"], ShouldBeNil)
				So(gc.rules["port"], ShouldBeNil)
				So(gc.rules["rate"], ShouldBeNil)
				So(gc.rules["nova"], ShouldBeNil)
			})

			Convey("potato is correct", func() {
				gc := cp.Get([]string{"one", "two", "potato"})
				So(gc, ShouldNotBeNil)

				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"].Required(), ShouldEqual, true)
				So(gc.rules["nova"].Required(), ShouldEqual, true)
				So(gc.rules["port"], ShouldNotBeNil)
				So(gc.rules["rate"], ShouldBeNil)
			})

			Convey("grapefruit is correct", func() {
				gc := cp.Get([]string{"one", "two", "grapefruit"})
				So(gc, ShouldNotBeNil)

				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"].Required(), ShouldEqual, true)
				So(gc.rules["nova"].Required(), ShouldEqual, true)
				So(gc.rules["port"], ShouldBeNil)
				So(gc.rules["rate"], ShouldNotBeNil)
			})

		})

	})
}
