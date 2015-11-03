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

package unpackage

import (
	"bufio"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnpackage(t *testing.T) {
	Convey("Valid gzipped tar", t, func() {
		infile := "pulse-collector-plugin-mock1.darwin-x86_64_gz.aci"

		path, m, err := Unpackager(infile)
		So(err, ShouldBeNil)
		So(path, ShouldContainSubstring, "pulse-collector-mock1")
		So(m, ShouldNotBeNil)
		So(m.App.Exec[0], ShouldContainSubstring, "pulse-collector-mock1")
	})
	Convey("Valid tar and manifest", t, func() {
		infile := "pulse-collector-plugin-mock1.darwin-x86_64.aci"

		f, err := os.Open(infile)
		defer f.Close()
		So(err, ShouldBeNil)

		data, err := Uncompress(infile, f)
		So(data, ShouldNotBeNil)
		So(err, ShouldBeNil)

		path, m, err := Unpackager(infile)
		So(err, ShouldBeNil)
		So(m, ShouldNotBeNil)
		So(path, ShouldContainSubstring, "pulse-collector-mock1")
		So(m.App.Exec[0], ShouldContainSubstring, "pulse-collector-mock1")
	})

	Convey("Plugin not .aci", t, func() {
		infile := "pulse-collector-plugin-mock1.darwin-x86_64/rootfs/pulse-collector-mock1"

		path, m, err := Unpackager(infile)
		So(err, ShouldBeNil)
		So(m, ShouldBeNil)
		So(path, ShouldContainSubstring, "pulse-collector-mock1")
	})

	Convey("Invalid tar (non existent file) and no data", t, func() {
		infile := "fakeFile.aci"

		f, err := os.Open(infile)
		defer f.Close()
		So(err, ShouldNotBeNil)

		data, err := Uncompress(infile, f)
		So(err, ShouldNotBeNil)
		So(data, ShouldBeNil)

		gr, err := Unzip(infile, f)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Error unzipping file")
		So(gr, ShouldBeNil)

		tr, err := Untar(infile, bufio.NewReader(f))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Error iterating through tar file")
		So(tr, ShouldBeNil)

		_, err = UnmarshalJSON(data)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Error unmarshaling JSON")

		path, m, err := Unpackager(infile)
		So(err, ShouldNotBeNil)
		So(path, ShouldResemble, "")
		So(m, ShouldBeNil)
	})
	Convey("Valid tar, manifest without exec file (plugin path)", t, func() {
		infile := "noExec.aci"
		path, m, err := Unpackager(infile)
		So(err, ShouldNotBeNil)
		So(path, ShouldResemble, "")
		So(m, ShouldBeNil)
	})
}
