package core

import (
	"testing"

	"github.com/intelsdilabs/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigDataTree(t *testing.T) {

	Convey("ConfigDataTree", t, func() {

		Convey("freeze should not panic", func() {
			cdt := NewTree()
			So(cdt, ShouldNotBeNil)
			So(cdt.cTree, ShouldNotBeNil)

			cd1 := NewNode()
			cd1.AddItem("s", ctypes.ConfigValueStr{Value: "foo"})
			cd1.AddItem("i", ctypes.ConfigValueInt{Value: -1})
			cd1.AddItem("f", ctypes.ConfigValueFloat{Value: -2.3})
			cd2 := NewNode()
			cd2.AddItem("s", ctypes.ConfigValueStr{Value: "bar"})
			cd2.AddItem("i", ctypes.ConfigValueInt{Value: 1})
			cd2.AddItem("f", ctypes.ConfigValueFloat{Value: 2.3})

			cdt.Add([]string{"1"}, cd1)
			cdt.Add([]string{"1", "2"}, cd2)

			So(func() {
				cdt.Freeze()
			}, ShouldNotPanic)
		})

		Convey("get", func() {
			cdt := NewTree()
			So(cdt, ShouldNotBeNil)
			So(cdt.cTree, ShouldNotBeNil)

			Convey("empty ns returns nil", func() {
				a := cdt.Get([]string{})
				So(a, ShouldBeNil)
			})

			Convey("validate complex tree", func() {
				cd1 := NewNode()
				cd1.AddItem("s", ctypes.ConfigValueStr{Value: "foo"})
				cd1.AddItem("x", ctypes.ConfigValueStr{Value: "wat"})
				cd1.AddItem("i", ctypes.ConfigValueInt{Value: -1})
				cd1.AddItem("f", ctypes.ConfigValueFloat{Value: -2.3})
				cd2 := NewNode()
				cd2.AddItem("s", ctypes.ConfigValueStr{Value: "bar"})
				cd2.AddItem("i", ctypes.ConfigValueInt{Value: 1})
				cd2.AddItem("f", ctypes.ConfigValueFloat{Value: 2.3})

				cdt.Add([]string{"1"}, cd1)
				cdt.Add([]string{"1", "2"}, cd2)

				a := cdt.Get([]string{"1", "2"})
				So(a, ShouldNotBeNil)

				t := a.Table()
				// This checks to ensure
				So(t["s"].Type(), ShouldEqual, "string")
				So(t["s"].(ctypes.ConfigValueStr).Value, ShouldEqual, "bar")
				So(t["x"].Type(), ShouldEqual, "string")
				So(t["x"].(ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
				So(t["i"].Type(), ShouldEqual, "integer")
				So(t["i"].(ctypes.ConfigValueInt).Value, ShouldEqual, 1)
				So(t["f"].Type(), ShouldEqual, "float")
				So(t["f"].(ctypes.ConfigValueFloat).Value, ShouldEqual, 2.3)
			})

		})

	})
}

func TestConfigDataNode(t *testing.T) {
	Convey("ConfigDataNode", t, func() {
		cd1 := NewNode()

		Convey("empty key adds no writes", func() {
			cd1.AddItem("", ctypes.ConfigValueStr{Value: "bar"})
			t := cd1.Table()
			So(len(t), ShouldEqual, 0)
		})

		Convey("adding a single string item exists in table", func() {
			cd1.AddItem("foo", ctypes.ConfigValueStr{Value: "bar"})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "string")
			So(t["foo"].(ctypes.ConfigValueStr).Value, ShouldEqual, "bar")
			So(len(t), ShouldEqual, 1)
		})

		Convey("adding a single int item exists in table", func() {
			cd1.AddItem("foo", ctypes.ConfigValueInt{Value: 1})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "integer")
			So(t["foo"].(ctypes.ConfigValueInt).Value, ShouldEqual, 1)
			So(len(t), ShouldEqual, 1)
		})

		Convey("adding a single float item exists in table", func() {
			cd1.AddItem("foo", ctypes.ConfigValueFloat{Value: 1.1})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "float")
			So(t["foo"].(ctypes.ConfigValueFloat).Value, ShouldEqual, 1.1)
			So(len(t), ShouldEqual, 1)
		})

		Convey("adding an item after another with the same key overwrites the value", func() {
			cd1.AddItem("foo", ctypes.ConfigValueFloat{Value: 1.1})
			cd1.AddItem("foo", ctypes.ConfigValueFloat{Value: 2.3})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "float")
			So(t["foo"].(ctypes.ConfigValueFloat).Value, ShouldEqual, 2.3)
			So(len(t), ShouldEqual, 1)
		})

		Convey("mutiple items added exist", func() {
			cd1.AddItem("s", ctypes.ConfigValueStr{Value: "bar"})
			cd1.AddItem("i", ctypes.ConfigValueInt{Value: 1})
			cd1.AddItem("f", ctypes.ConfigValueFloat{Value: 2.3})
			t := cd1.Table()
			So(t["s"].Type(), ShouldEqual, "string")
			So(t["s"].(ctypes.ConfigValueStr).Value, ShouldEqual, "bar")
			So(t["i"].Type(), ShouldEqual, "integer")
			So(t["i"].(ctypes.ConfigValueInt).Value, ShouldEqual, 1)
			So(t["f"].Type(), ShouldEqual, "float")
			So(t["f"].(ctypes.ConfigValueFloat).Value, ShouldEqual, 2.3)
			So(len(t), ShouldEqual, 3)
		})
	})
}
