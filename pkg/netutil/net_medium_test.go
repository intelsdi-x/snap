// +build medium

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

package netutil

import (
	"net"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetUtil(t *testing.T) {
	Convey("Get the first non-loopback ipv4 interface", t, func() {
		ip := GetIP()
		ifaces, err := net.Interfaces()
		So(err, ShouldBeNil)
		if len(ifaces) > 1 {
			So(ip, ShouldNotResemble, "127.0.0.1")
		} else {
			So(ip, ShouldResemble, "127.0.0.1")
		}
	})
}
