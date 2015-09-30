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
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"strings"
	"testing"
)

type IoctlArgsType struct {
	fd  uintptr
	cmd uint
	buf []byte
}

type OpenDeviceRetType struct {
	f *os.File
	e error
}

type fakeSysutilProvider struct {
	OpenDeviceArg []string
	IoctlArgs     []IoctlArgsType

	OpenDeviceRet OpenDeviceRetType
	IoctlRets     []error

	FillBuf []byte
}

func (s *fakeSysutilProvider) ListDevices() ([]string, error) {
	return []string{"DEV_ONE", "DEV_TWO"}, nil
}

func (s *fakeSysutilProvider) OpenDevice(device string) (*os.File, error) {
	s.OpenDeviceArg = append(s.OpenDeviceArg, device)

	return s.OpenDeviceRet.f, s.OpenDeviceRet.e
}

func (s *fakeSysutilProvider) Ioctl(fd uintptr, cmd uint, buf []byte) error {
	i := len(s.IoctlArgs)
	s.IoctlArgs = append(s.IoctlArgs, IoctlArgsType{fd, cmd, buf})

	//if

	return s.IoctlRets[i]
}

func firstKnownMetric() (byte, string) {
	for k, v := range AttributeMap {
		if !strings.Contains(v.Name, "/") {
			continue
		}
		return k, v.Name
	}

	panic("no attributes")
}

func TestSmartReader(t *testing.T) {

	Convey("Reading from smart capable device", t, func() {

		provider := &fakeSysutilProvider{OpenDeviceRet: OpenDeviceRetType{nil, nil},
			IoctlRets: []error{nil, nil}}
		_, err := ReadSmartData("MYDEV", provider)

		Convey("Should call OpenDevice from abstraction layer", func() {

			So(len(provider.OpenDeviceArg), ShouldEqual, 1)

		})

		Convey("Should call Ioctl twice", func() {

			So(provider.OpenDeviceArg[0], ShouldEqual, "MYDEV")

		})

		Convey("Should return no error", func() {

			So(err, ShouldBeNil)

		})

	})

	Convey("When it fails to open device during reading", t, func() {

		provider := &fakeSysutilProvider{OpenDeviceRet: OpenDeviceRetType{nil,
			errors.New("Something")},
			IoctlRets: []error{nil, nil}}
		_, err := ReadSmartData("MYDEV", provider)

		Convey("Should not call any ioctls", func() {

			So(len(provider.IoctlArgs), ShouldEqual, 0)

		})

		Convey("Should report error", func() {

			So(err, ShouldNotBeNil)

			Convey("Error should be about opening", func() {

				So(err.Error(), ShouldContainSubstring, "open")

			})

		})

	})

	Convey("When smart cannot be enabled during reading", t, func() {

		provider := &fakeSysutilProvider{OpenDeviceRet: OpenDeviceRetType{nil, nil},
			IoctlRets: []error{errors.New("Something"), nil}}
		_, err := ReadSmartData("MYDEV", provider)

		Convey("Should call ioctl once", func() {

			So(len(provider.IoctlArgs), ShouldEqual, 1)

		})

		Convey("Should report error", func() {

			So(err, ShouldNotBeNil)

			Convey("Error should be about enabling smart", func() {

				So(err.Error(), ShouldContainSubstring, "enable")

			})

		})

	})

	Convey("When device fails during reading but not during smart enabling phase",
		t, func() {

			provider := &fakeSysutilProvider{OpenDeviceRet: OpenDeviceRetType{nil, nil},
				IoctlRets: []error{nil, errors.New("Something else")}}
			_, err := ReadSmartData("MYDEV", provider)

			Convey("Should call ioctl twice", func() {

				So(len(provider.IoctlArgs), ShouldEqual, 2)

			})

			Convey("Should report error", func() {

				So(err, ShouldNotBeNil)

				Convey("Error should be about reading", func() {

					So(err.Error(), ShouldContainSubstring, "Read")

				})

			})

		})

}

func TestGetAttributes(t *testing.T) {
	Convey("When there is no known attribute", t, func() {

		sv := SmartValues{}
		sv.Values[0].Id = 255

		metrics := sv.GetAttributes()

		Convey("List of metrics should be empty", func() {

			So(metrics, ShouldBeEmpty)

		})

	})

	Convey("When known attribute is present", t, func() {

		id, value := firstKnownMetric()

		sv := SmartValues{}
		sv.Values[0].Id = id

		metrics := sv.GetAttributes()

		Convey("Should be present in list of metrics in raw form", func() {

			_, ok := metrics[value]
			So(ok, ShouldBeTrue)

		})

		Convey("Should be present in list of metrics in normalized form", func() {

			_, ok := metrics[value+"/normalized"]
			So(ok, ShouldBeTrue)

		})

	})
}

func TestAttributeFormat(t *testing.T) {

	format_desc := map[AttributeFormat]string{
		FormatDefault:     "FormatDefault",
		FormatFP1024:      "FormatFP1024",
		FormatPLPF:        "FormatPLPF",
		FormatTTS:         "FormatTTS",
		FormatTemperature: "FormatTemperature",
	}

	for format, label := range format_desc {
		Convey(fmt.Sprintf("Testing %v", label), t, func() {

			keys := format.GetKeys()

			Convey("Raw value is exposed", func() {

				So(keys, ShouldContain, "")

			})

			Convey("Normalized value is exposed", func() {

				So(keys, ShouldContain, "/normalized")

			})

			Convey("Keys and output of parsing is consistent", func() {

				raw := [8]byte{}
				parsed := format.ParseRaw(raw)
				Convey("Parsing returns every reported key", func() {
					for _, key := range keys {
						if key == "/normalized" {
							continue
						}
						_, ok := parsed[key]
						if !ok {
							Printf("%v is missing", key)
						}
						So(ok, ShouldBeTrue)
					}
				})
				Convey("Everything returned by parsing is reported", func() {
					for key, _ := range parsed {
						So(keys, ShouldContain, key)
					}
				})

			})
		})
	}
}
