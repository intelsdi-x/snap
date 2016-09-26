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

package cdata

import (
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/intelsdi-x/snap/core/ctypes"
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

				Convey("encode & decode", func() {
					gob.Register(&ConfigDataNode{})
					gob.Register(ctypes.ConfigValueStr{})
					gob.Register(ctypes.ConfigValueInt{})
					gob.Register(ctypes.ConfigValueFloat{})
					buf, err := cdt.GobEncode()
					So(err, ShouldBeNil)
					So(buf, ShouldNotBeNil)
					cdt2 := NewTree()
					err = cdt2.GobDecode(buf)
					So(err, ShouldBeNil)
					So(cdt2.cTree, ShouldNotBeNil)

					a2 := cdt2.Get([]string{"1", "2"})
					So(a2, ShouldNotBeNil)

					t2 := a2.Table()
					So(t2["s"].Type(), ShouldEqual, "string")
					println(fmt.Sprintf("!!! %T\n %T\n", t2["s"], t["s"]))
					So(t2["s"].(ctypes.ConfigValueStr).Value, ShouldEqual, "bar")
					So(t2["x"].Type(), ShouldEqual, "string")
					So(t2["x"].(ctypes.ConfigValueStr).Value, ShouldEqual, "wat")
					So(t2["i"].Type(), ShouldEqual, "integer")
					So(t2["i"].(ctypes.ConfigValueInt).Value, ShouldEqual, 1)
					So(t2["f"].Type(), ShouldEqual, "float")
					So(t2["f"].(ctypes.ConfigValueFloat).Value, ShouldEqual, 2.3)

				})
			})

		})

	})
}
