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

package main

import (
	"encoding/json"
	"testing"

	"github.com/intelsdi-x/snap/pkg/cfgfile"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapConfig(t *testing.T) {
	Convey("Test Config", t, func() {
		Convey("with defaults", func() {
			cfg := getDefaultConfig()
			jb, _ := json.Marshal(cfg)
			serrs := cfgfile.ValidateSchema(CONFIG_CONSTRAINTS, string(jb))
			So(len(serrs), ShouldEqual, 0)
		})
	})
}
