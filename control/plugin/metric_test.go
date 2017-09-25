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

package plugin

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMetric(t *testing.T) {
	Convey("error on invalid snap content type", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", 2),
		}
		a, c, e := MarshalMetricTypes("foo", m)
		m[0].Version_ = 1
		m[0].AddData(3)
		configNewNode := cdata.NewNode()
		configNewNode.AddItem("user", ctypes.ConfigValueStr{Value: "foo"})
		m[0].Config_ = configNewNode
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "invalid snap content type: foo")
		So(a, ShouldBeNil)
		So(c, ShouldEqual, "")
		So(m[0].Version(), ShouldResemble, 1)
		So(m[0].Data(), ShouldResemble, 3)
		So(m[0].Config(), ShouldNotBeNil)
	})

	Convey("error on empty metric slice", t, func() {
		m := []MetricType{}
		a, c, e := MarshalMetricTypes("foo", m)
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "attempt to marshall empty slice of metrics: foo")
		So(a, ShouldBeNil)
		So(c, ShouldEqual, "")
	})

	Convey("marshall using snap.* default to snap.gob", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.*", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.gob")

		Convey("unmarshal snap.gob", func() {
			m, e = UnmarshallMetricTypes("snap.gob", a)
			So(e, ShouldBeNil)
			So(m[0].Namespace().String(), ShouldResemble, "/foo/bar")
			So(m[0].Data(), ShouldResemble, 1)
			So(m[1].Namespace().String(), ShouldResemble, "/foo/baz")
			So(m[1].Data(), ShouldResemble, "2")
		})

	})

	Convey("marshall using snap.gob", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.gob", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.gob")

		Convey("unmarshal snap.gob", func() {
			m, e = UnmarshallMetricTypes("snap.gob", a)
			So(e, ShouldBeNil)
			So(m[0].Namespace().String(), ShouldResemble, "/foo/bar")
			So(m[0].Data(), ShouldResemble, 1)
			So(m[1].Namespace().String(), ShouldResemble, "/foo/baz")
			So(m[1].Data(), ShouldResemble, "2")
		})

		Convey("error on bad corrupt data", func() {
			a = []byte{1, 0, 1, 1, 1, 1, 1, 0, 0, 1}
			m, e = UnmarshallMetricTypes("snap.gob", a)
			So(e, ShouldNotBeNil)
			So(e.Error(), ShouldResemble, "gob: decoding into local type *[]plugin.MetricType, received remote type unknown type")
		})
	})

	Convey("marshall using snap.json", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.json")

		Convey("unmarshal snap.json", func() {
			m, e = UnmarshallMetricTypes("snap.json", a)
			So(e, ShouldBeNil)
			So(m[0].Namespace().String(), ShouldResemble, "/foo/bar")
			So(m[0].Data(), ShouldResemble, float64(1))
			So(m[1].Namespace().String(), ShouldResemble, "/foo/baz")
			So(m[1].Data(), ShouldResemble, "2")
		})

		Convey("error on bad corrupt data", func() {
			a = []byte{1, 0, 1, 1, 1, 1, 1, 0, 0, 1}
			m, e = UnmarshallMetricTypes("snap.json", a)
			So(e, ShouldNotBeNil)
			So(e.Error(), ShouldResemble, "invalid character '\\x01' looking for beginning of value")
		})
	})

	Convey("error on unmarshall using bad content type", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.json")

		m, e = UnmarshallMetricTypes("snap.wat", a)
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "invalid snap content type for unmarshalling: snap.wat")
		So(m, ShouldBeNil)
	})

	Convey("swap from snap.gob to snap.json", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.gob", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.gob")

		b, c, e := SwapMetricContentType(c, "snap.json", a)
		So(e, ShouldBeNil)
		So(c, ShouldResemble, "snap.json")
		So(b, ShouldNotBeNil)

		m, e = UnmarshallMetricTypes(c, b)
		So(e, ShouldBeNil)
		So(m[0].Namespace().String(), ShouldResemble, "/foo/bar")
		So(m[0].Data(), ShouldResemble, float64(1))
		So(m[1].Namespace().String(), ShouldResemble, "/foo/baz")
		So(m[1].Data(), ShouldResemble, "2")
	})

	Convey("swap from snap.json to snap.*", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.json")

		b, c, e := SwapMetricContentType(c, "snap.*", a)
		So(e, ShouldBeNil)
		So(c, ShouldResemble, "snap.gob")
		So(b, ShouldNotBeNil)

		m, e = UnmarshallMetricTypes(c, b)
		So(e, ShouldBeNil)
		So(m[0].Namespace().String(), ShouldResemble, "/foo/bar")
		So(m[0].Data(), ShouldResemble, float64(1))
		So(m[1].Namespace().String(), ShouldResemble, "/foo/baz")
		So(m[1].Data(), ShouldResemble, "2")
	})

	Convey("swap from snap.json to snap.gob", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.json")

		b, c, e := SwapMetricContentType(c, "snap.gob", a)
		So(e, ShouldBeNil)
		So(c, ShouldResemble, "snap.gob")
		So(b, ShouldNotBeNil)

		m, e = UnmarshallMetricTypes(c, b)
		So(e, ShouldBeNil)
		So(m[0].Namespace().String(), ShouldResemble, "/foo/bar")
		So(m[0].Data(), ShouldResemble, float64(1))
		So(m[1].Namespace().String(), ShouldResemble, "/foo/baz")
		So(m[1].Data(), ShouldResemble, "2")
	})

	Convey("error on bad content type to swap", t, func() {
		m := []MetricType{
			*NewMetricType(core.NewNamespace("foo", "bar"), time.Now(), nil, "", 1),
			*NewMetricType(core.NewNamespace("foo", "baz"), time.Now(), nil, "", "2"),
		}
		a, c, e := MarshalMetricTypes("snap.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "snap.json")

		b, c, e := SwapMetricContentType("snap.wat", "snap.gob", a)
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldResemble, "invalid snap content type for unmarshalling: snap.wat")
		So(c, ShouldEqual, "")
		So(b, ShouldBeNil)
	})
}
