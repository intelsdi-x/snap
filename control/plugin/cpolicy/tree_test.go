package cpolicy

import (
	"encoding/gob"
	"testing"

	"github.com/intelsdilabs/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigPolicyTree(t *testing.T) {
	Convey("ConfigPolicyTree", t, func() {
		t := NewTree()

		Convey("new tree", func() {
			So(t, ShouldNotBeNil)
		})

		Convey("stores a policy node", func() {
			cpn := NewPolicyNode()
			r1, _ := NewStringRule("username", false, "root")
			r2, _ := NewStringRule("password", true)
			cpn.Add(r1, r2)
			ns := []string{"one", "two", "potato"}

			t.Add(ns, cpn)
			t.Freeze()
			Convey("retrieves store policy", func() {
				gc := t.Get(ns)
				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["username"].Default().(*ctypes.ConfigValueStr).Value, ShouldEqual, "root")
				So(gc.rules["password"].Required(), ShouldEqual, true)
			})
			Convey("encode & decode", func() {
				gob.Register(NewPolicyNode())
				gob.Register(&stringRule{})
				buf, err := t.GobEncode()
				So(err, ShouldBeNil)
				So(buf, ShouldNotBeNil)
				t2 := NewTree()
				err = t2.GobDecode(buf)
				So(err, ShouldBeNil)
				So(t2.cTree, ShouldNotBeNil)
				gc := t2.Get([]string{"one", "two", "potato"})
				So(gc, ShouldNotBeNil)
				So(gc.rules["username"], ShouldNotBeNil)
				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"].Required(), ShouldEqual, true)
				So(gc.rules["username"].Default(), ShouldNotBeNil)
				So(gc.rules["password"].Default(), ShouldBeNil)
				So(gc.rules["username"].Default().(*ctypes.ConfigValueStr).Value, ShouldEqual, "root")
			})

		})

		Convey("stores multiple a policy nodes", func() {
			cpn1 := NewPolicyNode()
			r11, _ := NewStringRule("password", true)
			r12, _ := NewIntegerRule("port", true)
			cpn1.Add(r11, r12)
			ns1 := []string{"one", "two", "potato"}

			cpn2 := NewPolicyNode()
			r21, _ := NewStringRule("password", true)
			r22, _ := NewFloatRule("rate", true)
			cpn2.Add(r21, r22)
			ns2 := []string{"one", "two", "grapefruit"}

			cpn3 := NewPolicyNode()
			r31, _ := NewStringRule("username", false, "root")
			cpn3.Add(r31)
			ns3 := []string{"one", "two"}

			t.Add(ns1, cpn1)
			t.Add(ns2, cpn2)
			t.Add(ns3, cpn3)

			Convey("base node is nil", func() {
				gc := t.Get([]string{"one"})
				So(gc, ShouldBeNil)
			})

			Convey("two is correct", func() {
				gc := t.Get([]string{"one", "two"})
				So(gc, ShouldNotBeNil)

				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"], ShouldBeNil)
				So(gc.rules["port"], ShouldBeNil)
				So(gc.rules["rate"], ShouldBeNil)
			})

			Convey("potato is correct", func() {
				gc := t.Get([]string{"one", "two", "potato"})
				So(gc, ShouldNotBeNil)

				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"].Required(), ShouldEqual, true)
				So(gc.rules["port"], ShouldNotBeNil)
				So(gc.rules["rate"], ShouldBeNil)
			})

			Convey("grapefruit is correct", func() {
				gc := t.Get([]string{"one", "two", "grapefruit"})
				So(gc, ShouldNotBeNil)

				So(gc.rules["username"].Required(), ShouldEqual, false)
				So(gc.rules["password"].Required(), ShouldEqual, true)
				So(gc.rules["port"], ShouldBeNil)
				So(gc.rules["rate"], ShouldNotBeNil)
			})

		})

	})
}
