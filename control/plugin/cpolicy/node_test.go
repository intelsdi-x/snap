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
	"testing"

	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

// Helpful method to switch to slice of strings for goconvey contains checking
func errorsMsg(errors []error) []string {
	s := []string{}
	for _, e := range errors {
		s = append(s, e.Error())
	}
	return s
}

func TestConfigPolicyNode(t *testing.T) {

	Convey("returns error for missing required data", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["username"] = ctypes.ConfigValueStr{Value: "root"}

		r1, _ := NewStringRule("username", true)
		r2, _ := NewStringRule("password", true)
		r3, _ := NewBoolRule("nova", false)

		n.Add(r1, r2, r3)

		_, pe := n.Process(m)

		So(pe.HasErrors(), ShouldBeTrue)
		So(len(pe.Errors()), ShouldEqual, 1)
		So(errorsMsg(pe.Errors()), ShouldContain, "required key missing (password)")
	})

	Convey("returns errors for missing required data (multiple)", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["username"] = ctypes.ConfigValueStr{Value: "root"}

		r1, _ := NewStringRule("username", true)
		r2, _ := NewStringRule("password", true)
		r3, _ := NewIntegerRule("port", true)
		r4, _ := NewIntegerRule("port", true)

		n.Add(r1, r2, r3, r4)

		_, pe := n.Process(m)

		So(pe.HasErrors(), ShouldBeTrue)
		So(len(pe.Errors()), ShouldEqual, 2)
		So(errorsMsg(pe.Errors()), ShouldContain, "required key missing (port)")
		So(errorsMsg(pe.Errors()), ShouldContain, "required key missing (password)")
	})

	Convey("returns error for mismatched type", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["username"] = ctypes.ConfigValueStr{Value: "root"}
		m["password"] = ctypes.ConfigValueStr{Value: "password"}
		m["port"] = ctypes.ConfigValueStr{Value: "8080"}
		m["nova"] = ctypes.ConfigValueStr{Value: "true"}

		r1, _ := NewStringRule("username", true)
		r2, _ := NewStringRule("password", true)
		r3, _ := NewIntegerRule("port", true)
		r4, _ := NewBoolRule("nova", true)

		n.Add(r1, r2, r3, r4)

		_, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 2)
		So(errorsMsg(pe.Errors()), ShouldContain, "type mismatch (port wanted type 'integer' but provided type 'string')")
		So(errorsMsg(pe.Errors()), ShouldContain, "type mismatch (nova wanted type 'bool' but provided type 'string')")
	})

	Convey("adds defaults to only missing values that should have them", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["password"] = ctypes.ConfigValueStr{Value: "password"}

		r1, _ := NewStringRule("username", false, "root")
		r2, _ := NewStringRule("password", true)
		r3, _ := NewIntegerRule("port", false, 8080)
		r4, _ := NewBoolRule("nova", false, true)

		n.Add(r1, r2, r3, r4)

		m2, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 0)
		So((*m2)["username"].(ctypes.ConfigValueStr).Value, ShouldEqual, "root")
		So((*m2)["port"].(ctypes.ConfigValueInt).Value, ShouldEqual, 8080)
		So((*m2)["nova"].(ctypes.ConfigValueBool).Value, ShouldEqual, true)
	})

	Convey("defaults don't fix missing values on required", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["password"] = ctypes.ConfigValueStr{Value: "password"}

		r1, _ := NewStringRule("username", true, "root")
		r2, _ := NewStringRule("password", true)

		n.Add(r1, r2)

		m2, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 1)
		So(m2, ShouldBeNil)
	})

	Convey("Test integer value between minimum and maximum for rule succeeds", t, func() {
		n := NewPolicyNode()

		m := map[string]ctypes.ConfigValue{}
		m["port"] = ctypes.ConfigValueInt{Value: 5}

		r1, _ := NewIntegerRule("port", false)
		r1.SetMinimum(0)
		r1.SetMaximum(10)

		n.Add(r1)

		m2, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 0)
		So((*m2)["port"].(ctypes.ConfigValueInt).Value, ShouldEqual, 5)
	})
	Convey("Test integer value outside of maximum for rule fails", t, func() {
		n := NewPolicyNode()

		m := map[string]ctypes.ConfigValue{}
		m["port"] = ctypes.ConfigValueInt{Value: 5}

		r1, _ := NewIntegerRule("port", false)
		r1.SetMinimum(0)
		r1.SetMaximum(4)

		n.Add(r1)

		_, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 1)
	})
	Convey("Test integer value outside of minimum for rule fails", t, func() {
		n := NewPolicyNode()

		m := map[string]ctypes.ConfigValue{}
		m["port"] = ctypes.ConfigValueInt{Value: 5}

		r1, _ := NewIntegerRule("port", false)
		r1.SetMinimum(7)
		r1.SetMaximum(100)

		n.Add(r1)

		_, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 1)
	})
	Convey("Test float value between minimum and maximum for rule succeeds", t, func() {
		n := NewPolicyNode()

		m := map[string]ctypes.ConfigValue{}
		m["num"] = ctypes.ConfigValueFloat{Value: 5.06}

		r1, _ := NewFloatRule("num", false)
		r1.SetMinimum(3.14)
		r1.SetMaximum(10.17)

		n.Add(r1)

		m2, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 0)
		So((*m2)["num"].(ctypes.ConfigValueFloat).Value, ShouldEqual, 5.06)
	})
	Convey("Test float value outside of maximum for rule fails", t, func() {
		n := NewPolicyNode()

		m := map[string]ctypes.ConfigValue{}
		m["num"] = ctypes.ConfigValueFloat{Value: 5.97}

		r1, _ := NewFloatRule("num", false)
		r1.SetMinimum(0)
		r1.SetMaximum(3.14)

		n.Add(r1)

		_, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 1)
	})
	Convey("Test float value outside of minimum for rule fails", t, func() {
		n := NewPolicyNode()

		m := map[string]ctypes.ConfigValue{}
		m["num"] = ctypes.ConfigValueFloat{Value: 2.26}

		r1, _ := NewFloatRule("num", false)
		r1.SetMinimum(3.14)
		r1.SetMaximum(42)

		n.Add(r1)

		_, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 1)
	})

}
