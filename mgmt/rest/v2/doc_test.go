// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package v2

import (
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/go-swagger/go-swagger/scan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {
	Convey("Swagger spec validation", t, func() {
		doc, err := loads.Spec("swagger.json")
		Convey("Open current swagger.json", func() {
			So(err, ShouldBeNil)
			So(doc, ShouldNotBeNil)
		})
		err = validate.Spec(doc, strfmt.Default)
		Convey("Validate local spec", func() {
			So(err, ShouldBeNil)
		})

		opts := scan.Opts{
			BasePath: ".",
		}
		spec, err := scan.Application(opts)
		Convey("Generate spec on the current project", func() {
			So(err, ShouldBeNil)
			So(spec, ShouldNotBeNil)
		})

		json, err := doc.Spec().MarshalJSON()
		json2, err2 := spec.MarshalJSON()
		Convey("Compare local spec with generated spec", func() {
			So(err, ShouldBeNil)
			So(err2, ShouldBeNil)
			So(string(json), ShouldEqual, string(json2))
		})
	})
}
