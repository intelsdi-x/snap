/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package smart

import (
	"errors"
	"github.com/intelsdi-x/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"strings"
	"testing"
)

type fakeSysutilProvider2 struct {
	FillBuf []byte
}

func (s *fakeSysutilProvider2) ListDevices() ([]string, error) {
	return []string{"DEV_ONE", "DEV_TWO"}, nil
}

func (s *fakeSysutilProvider2) OpenDevice(device string) (*os.File, error) {
	return nil, nil
}

func (s *fakeSysutilProvider2) Ioctl(fd uintptr, cmd uint, buf []byte) error {
	if cmd == smart_read_values {
		for i, v := range s.FillBuf {
			buf[i] = v
		}
	}
	return nil
}

func sysUtilWithMetrics(metrics []byte) fakeSysutilProvider2 {
	util := fakeSysutilProvider2{FillBuf: make([]byte, 512)}

	for i, m := range metrics {
		util.FillBuf[2+i*12] = m
	}

	return util
}

func TestGetMetricTypes(t *testing.T) {
	Convey("When having two devices with known smart attribute", t, func() {

		Convey("And system lets you to list devices", func() {
			provider := &fakeSysutilProvider2{}

			orgProvider := sysUtilProvider
			sysUtilProvider = provider

			collector := SmartCollector{}

			Convey("Both devices should be present in metric list", func() {

				dev_one, dev_two := false, false
				metrics, _ := collector.GetMetricTypes()

				for _, m := range metrics {
					switch m.Namespace()[2] {
					case "DEV_ONE":
						dev_one = true
					case "DEV_TWO":
						dev_two = true
					}
				}

				So(dev_one, ShouldBeTrue)
				So(dev_two, ShouldBeTrue)

			})

			Reset(func() {
				sysUtilProvider = orgProvider
			})

		})

	})

}

func TestParseName(t *testing.T) {
	Convey("When given correct namespace refering to single word attribute", t, func() {

		disk, attr := parseName([]string{"intel", "disk", "DEV", "smart", "abc"})

		Convey("Device should be correctly extracted", func() {

			So(disk, ShouldEqual, "DEV")

		})

		Convey("Attribute should be correctly extracted", func() {

			So(attr, ShouldEqual, "abc")

		})

	})

	Convey("When given correct namespace refering to multi level attribute", t, func() {

		disk, attr := parseName([]string{"intel", "disk", "DEV", "smart",
			"abc", "def"})

		Convey("Device should be correctly extracted", func() {

			So(disk, ShouldEqual, "DEV")

		})

		Convey("Attribute should be correctly extracted", func() {

			So(attr, ShouldEqual, "abc/def")

		})

	})

}

func TestValidateName(t *testing.T) {
	Convey("When given namespace with invalid prefix", t, func() {

		test := validateName([]string{"intel", "cake", "DEV", "smart",
			"abc", "def"})

		Convey("Validation should fail", func() {

			So(test, ShouldBeFalse)

		})

	})

	Convey("When given namespace with invalid suffix", t, func() {

		test := validateName([]string{"intel", "disk", "DEV", "dumb",
			"abc", "def"})

		Convey("Validation should fail", func() {

			So(test, ShouldBeFalse)

		})

	})

	Convey("When given correct namespace refering to single word attribute", t, func() {

		test := validateName([]string{"intel", "disk", "DEV", "smart", "abc"})

		Convey("Validation should pass", func() {

			So(test, ShouldBeTrue)

		})

	})

	Convey("When given correct namespace refering to multi level attribute", t, func() {

		test := validateName([]string{"intel", "disk", "DEV", "smart",
			"abc", "def"})
		Convey("Validation should pass", func() {

			So(test, ShouldBeTrue)

		})

	})
}

func TestCollectMetrics(t *testing.T) {
	Convey("Using fake system", t, func() {

		orgReader := ReadSmartData
		orgProvider := sysUtilProvider

		sc := SmartCollector{}

		metric_id, metric_name := firstKnownMetric()
		metric_ns := strings.Split(metric_name, "/")

		Convey("When asked about metric not in valid namespace", func() {

			_, err := sc.CollectMetrics([]plugin.PluginMetricType{
				{Namespace_: []string{"cake"}}})

			Convey("Returns error", func() {

				So(err, ShouldNotBeNil)

				Convey("Error is about invalid metric", func() {

					So(err.Error(), ShouldContainSubstring, "not valid")

				})

			})

		})

		Convey("When asked about metric in valid namespace but unknown to reader", func() {

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				return &SmartValues{}, nil
			}

			_, err := sc.CollectMetrics([]plugin.PluginMetricType{
				{Namespace_: []string{"intel", "disk", "x", "smart", "y"}}})

			Convey("Returns error", func() {

				So(err, ShouldNotBeNil)

				Convey("Error is about unknown metric", func() {

					So(err.Error(), ShouldContainSubstring, "Unknown")

				})

			})

		})

		Convey("When asked about metric in valid namespace but reading fails", func() {

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				return nil, errors.New("Something")
			}

			_, err := sc.CollectMetrics([]plugin.PluginMetricType{
				{Namespace_: []string{"intel", "disk", "x", "smart", "y"}}})

			Convey("Returns error", func() {

				So(err, ShouldNotBeNil)

			})

		})

		Convey("When asked about metric in valid namespace", func() {

			drive_asked := ""

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				drive_asked = device

				result := SmartValues{}
				result.Values[0].Id = metric_id

				return &result, nil
			}

			metrics, _ := sc.CollectMetrics([]plugin.PluginMetricType{
				{Namespace_: append([]string{"intel", "disk", "x", "smart"},
					metric_ns...)}})

			Convey("Asks reader to read metric from correct drive", func() {

				So(drive_asked, ShouldEqual, "x")

				Convey("Returns value of metric from reader", func() {
					So(len(metrics), ShouldBeGreaterThan, 0)

					//TODO: Value is correct

				})

			})

		})

		Convey("When asked about metrics in valid namespaces", func() {

			asked := map[string]int{"x": 0, "y": 0}

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				asked[device]++

				result := SmartValues{}
				result.Values[0].Id = metric_id

				return &result, nil
			}
			sc.CollectMetrics([]plugin.PluginMetricType{
				{Namespace_: append([]string{"intel", "disk", "x", "smart"}, metric_ns...)},
				{Namespace_: append([]string{"intel", "disk", "y", "smart"}, metric_ns...)},
				{Namespace_: append([]string{"intel", "disk", "y", "smart"}, metric_ns...)},
				{Namespace_: append([]string{"intel", "disk", "x", "smart"}, metric_ns...)},
			})

			Convey("Reader is asked once per drive", func() {
				So(asked["x"], ShouldEqual, 1)
				So(asked["y"], ShouldEqual, 1)

			})

		})

		Reset(func() {
			sysUtilProvider = orgProvider
			ReadSmartData = orgReader
		})

	})
}
