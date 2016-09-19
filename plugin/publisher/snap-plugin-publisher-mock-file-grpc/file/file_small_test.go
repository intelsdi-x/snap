// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilePublish(t *testing.T) {
	metrics := []plugin.Metric{plugin.Metric{Namespace: plugin.NewNamespace("foo"), Timestamp: time.Now(), Data: 99, Tags: nil}}
	config := plugin.Config{"file": "/tmp/pub.out"}
	Convey("TestFilePublish", t, func() {
		fp := NewFilePublisher()
		So(fp, ShouldNotBeNil)
		err := fp.Publish(metrics, config)
		So(err, ShouldBeNil)
		filename, _ := config.GetString("file")
		_, err = os.Stat(filename)
		So(err, ShouldBeNil)
	})
}
