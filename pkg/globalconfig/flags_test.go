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

package globalconfig

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFlags(t *testing.T) {
	Convey("Provided a config in JSON we are able to unmarshal it into a valid config", t, func() {
		cfg := NewConfig()
		cfg.LoadConfig("../../examples/configs/snap-config-sample.json")
		So(cfg.Flags, ShouldNotBeNil)
		So(cfg.Flags.LogLevel, ShouldNotBeNil)
		So(cfg.Flags.PluginTrust, ShouldNotBeNil)
		So(cfg.Flags.AutodiscoverPath, ShouldNotBeNil)
		So(*cfg.Flags.LogLevel, ShouldEqual, 1)
		So(*cfg.Flags.PluginTrust, ShouldEqual, 0)
		So(*cfg.Flags.AutodiscoverPath, ShouldEqual, "build/plugin")
	})
}
