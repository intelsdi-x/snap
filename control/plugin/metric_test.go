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
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMetric(t *testing.T) {
	Convey("error on invalid pulse content type", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, 2),
		}
		a, c, e := MarshalPluginMetricTypes("foo", m)
		m[0].Version_ = 1
		m[0].AddData(3)
		configNewNode := cdata.NewNode()
		configNewNode.AddItem("user", ctypes.ConfigValueStr{Value: "foo"})
		m[0].Config_ = configNewNode
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "invalid pulse content type: foo")
		So(a, ShouldBeNil)
		So(c, ShouldEqual, "")
		So(m[0].Version(), ShouldResemble, 1)
		So(m[0].Data(), ShouldResemble, 3)
		So(m[0].Config(), ShouldNotBeNil)
	})

	Convey("error on empty metric slice", t, func() {
		m := []PluginMetricType{}
		a, c, e := MarshalPluginMetricTypes("foo", m)
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "attempt to marshall empty slice of metrics: foo")
		So(a, ShouldBeNil)
		So(c, ShouldEqual, "")
	})

	Convey("marshall using pulse.* default to pulse.gob", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.*", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.gob")

		Convey("unmarshal pulse.gob", func() {
			m, e = UnmarshallPluginMetricTypes("pulse.gob", a)
			So(e, ShouldBeNil)
			So(strings.Join(m[0].Namespace(), "/"), ShouldResemble, "foo/bar")
			So(m[0].Data(), ShouldResemble, 1)
			So(strings.Join(m[1].Namespace(), "/"), ShouldResemble, "foo/baz")
			So(m[1].Data(), ShouldResemble, "2")
		})

	})

	Convey("marshall using pulse.gob", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.gob", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.gob")

		Convey("unmarshal pulse.gob", func() {
			m, e = UnmarshallPluginMetricTypes("pulse.gob", a)
			So(e, ShouldBeNil)
			So(strings.Join(m[0].Namespace(), "/"), ShouldResemble, "foo/bar")
			So(m[0].Data(), ShouldResemble, 1)
			So(strings.Join(m[1].Namespace(), "/"), ShouldResemble, "foo/baz")
			So(m[1].Data(), ShouldResemble, "2")
		})

		Convey("error on bad corrupt data", func() {
			a = []byte{1, 0, 1, 1, 1, 1, 1, 0, 0, 1}
			m, e = UnmarshallPluginMetricTypes("pulse.gob", a)
			So(e, ShouldNotBeNil)
			So(e.Error(), ShouldResemble, "gob: decoding into local type *[]plugin.PluginMetricType, received remote type unknown type")
		})
	})

	Convey("marshall using pulse.json", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.json")

		Convey("unmarshal pulse.json", func() {
			m, e = UnmarshallPluginMetricTypes("pulse.json", a)
			So(e, ShouldBeNil)
			So(strings.Join(m[0].Namespace(), "/"), ShouldResemble, "foo/bar")
			So(m[0].Data(), ShouldResemble, float64(1))
			So(strings.Join(m[1].Namespace(), "/"), ShouldResemble, "foo/baz")
			So(m[1].Data(), ShouldResemble, "2")
		})

		Convey("error on bad corrupt data", func() {
			a = []byte{1, 0, 1, 1, 1, 1, 1, 0, 0, 1}
			m, e = UnmarshallPluginMetricTypes("pulse.json", a)
			So(e, ShouldNotBeNil)
			So(e.Error(), ShouldResemble, "invalid character '\\x01' looking for beginning of value")
		})
	})

	Convey("error on unmarshall using bad content type", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.json")

		m, e = UnmarshallPluginMetricTypes("pulse.wat", a)
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "invalid pulse content type for unmarshalling: pulse.wat")
		So(m, ShouldBeNil)
	})

	Convey("swap from pulse.gob to pulse.json", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.gob", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.gob")

		b, c, e := SwapPluginMetricContentType(c, "pulse.json", a)
		So(e, ShouldBeNil)
		So(c, ShouldResemble, "pulse.json")
		So(b, ShouldNotBeNil)

		m, e = UnmarshallPluginMetricTypes(c, b)
		So(e, ShouldBeNil)
		So(strings.Join(m[0].Namespace(), "/"), ShouldResemble, "foo/bar")
		So(m[0].Data(), ShouldResemble, float64(1))
		So(strings.Join(m[1].Namespace(), "/"), ShouldResemble, "foo/baz")
		So(m[1].Data(), ShouldResemble, "2")
	})

	Convey("swap from pulse.json to pulse.*", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.json")

		b, c, e := SwapPluginMetricContentType(c, "pulse.*", a)
		So(e, ShouldBeNil)
		So(c, ShouldResemble, "pulse.gob")
		So(b, ShouldNotBeNil)

		m, e = UnmarshallPluginMetricTypes(c, b)
		So(e, ShouldBeNil)
		So(strings.Join(m[0].Namespace(), "/"), ShouldResemble, "foo/bar")
		So(m[0].Data(), ShouldResemble, float64(1))
		So(strings.Join(m[1].Namespace(), "/"), ShouldResemble, "foo/baz")
		So(m[1].Data(), ShouldResemble, "2")
	})

	Convey("swap from pulse.json to pulse.gob", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.json")

		b, c, e := SwapPluginMetricContentType(c, "pulse.gob", a)
		So(e, ShouldBeNil)
		So(c, ShouldResemble, "pulse.gob")
		So(b, ShouldNotBeNil)

		m, e = UnmarshallPluginMetricTypes(c, b)
		So(e, ShouldBeNil)
		So(strings.Join(m[0].Namespace(), "/"), ShouldResemble, "foo/bar")
		So(m[0].Data(), ShouldResemble, float64(1))
		So(strings.Join(m[1].Namespace(), "/"), ShouldResemble, "foo/baz")
		So(m[1].Data(), ShouldResemble, "2")
	})

	Convey("error on bad content type to swap", t, func() {
		m := []PluginMetricType{
			*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", nil, nil, 1),
			*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", nil, nil, "2"),
		}
		a, c, e := MarshalPluginMetricTypes("pulse.json", m)
		So(e, ShouldBeNil)
		So(a, ShouldNotBeNil)
		So(len(a), ShouldBeGreaterThan, 0)
		So(c, ShouldEqual, "pulse.json")

		b, c, e := SwapPluginMetricContentType("pulse.wat", "pulse.gob", a)
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldResemble, "invalid pulse content type for unmarshalling: pulse.wat")
		So(b, ShouldBeNil)
	})
}
