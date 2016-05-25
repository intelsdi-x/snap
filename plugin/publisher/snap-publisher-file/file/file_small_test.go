// +build small

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

package file

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilePublish(t *testing.T) {
	var buf bytes.Buffer
	metrics := []plugin.MetricType{
		*plugin.NewMetricType(core.NewNamespace("foo"), time.Now(), nil, "", 99),
	}
	config := make(map[string]ctypes.ConfigValue)
	enc := gob.NewEncoder(&buf)
	enc.Encode(metrics)

	Convey("TestFilePublish", t, func() {
		config["file"] = ctypes.ConfigValueStr{Value: "/tmp/pub.out"}
		fp := NewFilePublisher()
		So(fp, ShouldNotBeNil)
		err := fp.Publish("", buf.Bytes(), config)
		So(err, ShouldResemble, errors.New("Unknown content type ''"))
		err = fp.Publish(plugin.SnapGOBContentType, buf.Bytes(), config)
		So(err, ShouldBeNil)
		_, err = os.Stat(config["file"].(ctypes.ConfigValueStr).Value)
		So(err, ShouldBeNil)
		meta := Meta()
		So(meta, ShouldNotBeNil)
	})
}
