package cpolicy

import (
	"testing"

	"github.com/intelsdi-x/pulse/core/ctypes"
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

		n.Add(r1, r2)

		_, pe := n.Process(m)

		So(pe.HasErrors(), ShouldBeTrue)
		So(len(pe.Errors()), ShouldEqual, 1)
		So(errorsMsg(pe.Errors()), ShouldContain, "required key missing (password)")
	})

	Convey("returns errors for missing required data (mutliple)", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["username"] = ctypes.ConfigValueStr{Value: "root"}

		r1, _ := NewStringRule("username", true)
		r2, _ := NewStringRule("password", true)
		r3, _ := NewIntegerRule("port", true)

		n.Add(r1, r2, r3)

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

		r1, _ := NewStringRule("username", true)
		r2, _ := NewStringRule("password", true)
		r3, _ := NewIntegerRule("port", true)

		n.Add(r1, r2, r3)

		_, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 1)
		So(errorsMsg(pe.Errors()), ShouldContain, "type mismatch (port wanted type 'integer' but provided type 'string')")
	})

	Convey("adds defaults to only missing values that should have them", t, func() {
		n := NewPolicyNode()
		So(n, ShouldNotBeNil)

		m := map[string]ctypes.ConfigValue{}
		m["password"] = ctypes.ConfigValueStr{Value: "password"}

		r1, _ := NewStringRule("username", false, "root")
		r2, _ := NewStringRule("password", true)
		r3, _ := NewIntegerRule("port", false, 8080)

		n.Add(r1, r2, r3)

		m2, pe := n.Process(m)

		So(len(pe.Errors()), ShouldEqual, 0)
		So((*m2)["username"].(*ctypes.ConfigValueStr).Value, ShouldEqual, "root")
		So((*m2)["port"].(*ctypes.ConfigValueInt).Value, ShouldEqual, 8080)
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

}
