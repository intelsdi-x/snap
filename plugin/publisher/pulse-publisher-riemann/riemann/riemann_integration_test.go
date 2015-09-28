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

package riemann

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/amir/raidman"
)

// integration test
func TestPublish(t *testing.T) {
	// This integration test requires a Riemann Server
	broker := os.Getenv("PULSE_TEST_RIEMANN")
	if broker == "" {
		fmt.Println("Skipping integration tests")
		return
	}

	Convey("Publish to Riemann", t, func() {
		Convey("publish and consume", func() {
			r := NewRiemannPublisher()
			cdn := cdata.NewNode()
			cdn.AddItem("broker", ctypes.ConfigValueStr{Value: broker})
			cdn.AddItem("host", ctypes.ConfigValueStr{Value: "bacon-powered"})
			cp := r.GetConfigPolicy()
			p := cp.Get([]string{""})
			f, cErr := p.Process(cdn.Table())
			So(getProcessErrorStr(cErr), ShouldEqual, "")

			metrics := []plugin.PluginMetricType{
				*plugin.NewPluginMetricType([]string{"intel", "cpu", "temp"}, time.Now(), "", 100),
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			err := r.Publish(plugin.PulseGOBContentType, buf.Bytes(), *f)
			So(err, ShouldBeNil)

			c, _ := raidman.Dial("tcp", broker)
			events, _ := c.Query("host = \"bacon-powered\"")
			So(len(events), ShouldBeGreaterThan, 0)
		})
	})
}

func getProcessErrorStr(err *cpolicy.ProcessingErrors) string {
	errString := ""
	if err.HasErrors() {
		for _, e := range err.Errors() {
			errString += fmt.Sprintln(e.Error())
		}
	}
	return errString
}
