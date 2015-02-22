package cdata

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
