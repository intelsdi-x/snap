package core

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigDataTree(t *testing.T) {

	Convey("ConfigDataTree", t, func() {

		Convey("freeze should not panic", func() {
			cdt := NewConfigDataTree()
			So(cdt, ShouldNotBeNil)
			So(cdt.cTree, ShouldNotBeNil)

			cd1 := NewConfigDataNode()
			cd1.AddItem("s", configValueStr{Value: "foo"})
			cd1.AddItem("i", configValueInt{Value: -1})
			cd1.AddItem("f", configValueFloat{Value: -2.3})
			cd2 := NewConfigDataNode()
			cd2.AddItem("s", configValueStr{Value: "bar"})
			cd2.AddItem("i", configValueInt{Value: 1})
			cd2.AddItem("f", configValueFloat{Value: 2.3})

			cdt.Add([]string{"1"}, cd1)
			cdt.Add([]string{"1", "2"}, cd2)

			So(func() {
				cdt.Freeze()
			}, ShouldNotPanic)
		})

		Convey("get", func() {
			cdt := NewConfigDataTree()
			So(cdt, ShouldNotBeNil)
			So(cdt.cTree, ShouldNotBeNil)

			Convey("empty ns returns nil", func() {
				a := cdt.Get([]string{})
				So(a, ShouldBeNil)
			})

			Convey("validate complex tree", func() {
				cd1 := NewConfigDataNode()
				cd1.AddItem("s", configValueStr{Value: "foo"})
				cd1.AddItem("x", configValueStr{Value: "wat"})
				cd1.AddItem("i", configValueInt{Value: -1})
				cd1.AddItem("f", configValueFloat{Value: -2.3})
				cd2 := NewConfigDataNode()
				cd2.AddItem("s", configValueStr{Value: "bar"})
				cd2.AddItem("i", configValueInt{Value: 1})
				cd2.AddItem("f", configValueFloat{Value: 2.3})

				cdt.Add([]string{"1"}, cd1)
				cdt.Add([]string{"1", "2"}, cd2)

				a := cdt.Get([]string{"1", "2"})
				So(a, ShouldNotBeNil)

				t := a.Table()
				// This checks to ensure
				So(t["s"].Type(), ShouldEqual, "string")
				So(t["s"].(configValueStr).Value, ShouldEqual, "bar")
				So(t["x"].Type(), ShouldEqual, "string")
				So(t["x"].(configValueStr).Value, ShouldEqual, "wat")
				So(t["i"].Type(), ShouldEqual, "integer")
				So(t["i"].(configValueInt).Value, ShouldEqual, 1)
				So(t["f"].Type(), ShouldEqual, "float")
				So(t["f"].(configValueFloat).Value, ShouldEqual, 2.3)
			})

		})

	})
}

func TestConfigDataNode(t *testing.T) {
	Convey("ConfigDataNode", t, func() {
		cd1 := NewConfigDataNode()

		Convey("empty key adds no writes", func() {
			cd1.AddItem("", configValueStr{Value: "bar"})
			t := cd1.Table()
			So(len(t), ShouldEqual, 0)
		})

		Convey("adding a single string item exists in table", func() {
			cd1.AddItem("foo", configValueStr{Value: "bar"})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "string")
			So(t["foo"].(configValueStr).Value, ShouldEqual, "bar")
			So(len(t), ShouldEqual, 1)
		})

		Convey("adding a single int item exists in table", func() {
			cd1.AddItem("foo", configValueInt{Value: 1})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "integer")
			So(t["foo"].(configValueInt).Value, ShouldEqual, 1)
			So(len(t), ShouldEqual, 1)
		})

		Convey("adding a single float item exists in table", func() {
			cd1.AddItem("foo", configValueFloat{Value: 1.1})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "float")
			So(t["foo"].(configValueFloat).Value, ShouldEqual, 1.1)
			So(len(t), ShouldEqual, 1)
		})

		Convey("adding an item after another with the same key overwrites the value", func() {
			cd1.AddItem("foo", configValueFloat{Value: 1.1})
			cd1.AddItem("foo", configValueFloat{Value: 2.3})
			t := cd1.Table()
			So(t["foo"].Type(), ShouldEqual, "float")
			So(t["foo"].(configValueFloat).Value, ShouldEqual, 2.3)
			So(len(t), ShouldEqual, 1)
		})

		Convey("mutiple items added exist", func() {
			cd1.AddItem("s", configValueStr{Value: "bar"})
			cd1.AddItem("i", configValueInt{Value: 1})
			cd1.AddItem("f", configValueFloat{Value: 2.3})
			t := cd1.Table()
			So(t["s"].Type(), ShouldEqual, "string")
			So(t["s"].(configValueStr).Value, ShouldEqual, "bar")
			So(t["i"].Type(), ShouldEqual, "integer")
			So(t["i"].(configValueInt).Value, ShouldEqual, 1)
			So(t["f"].Type(), ShouldEqual, "float")
			So(t["f"].(configValueFloat).Value, ShouldEqual, 2.3)
			So(len(t), ShouldEqual, 3)
		})
	})
}
