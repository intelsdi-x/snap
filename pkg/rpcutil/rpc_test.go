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

package rpcutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRpcUtil(t *testing.T) {
	ca := "../../examples/certs/sample-ca.pem"
	caKey := "../../examples/certs/sample-ca-key.pem"
	signedCert := "../../examples/certs/sample-signed-cert.pem"
	signedKey := "../../examples/certs/sample-signed-cert-key.pem"
	Convey("Get a grpc.ServerOption enabling tls", t, func() {
		opt, err := ServerTlsOption(ca, caKey, "127.0.0.1")
		Convey("Provided a valid root cert, key and listen address", func() {
			So(err, ShouldBeNil)
			So(opt, ShouldNotBeNil)
		})
		opt, err = ServerTlsOption(signedCert, signedKey, "127.0.0.1")
		Convey("Provided a cert that is not a CA", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrCertTrust)
			So(opt, ShouldBeNil)
		})
		opt, err = ServerTlsOption(ca, caKey, "1")
		Convey("Provided an invalid listen address", func() {
			So(err, ShouldNotBeNil)
			So(opt, ShouldBeNil)
		})
	})

	Convey("Get a client connection", t, func() {
		conn, err := GetClientConnection("127.0.0.1", 8183, ca, caKey)
		Convey("Provided a valid address, port, root cert and key", func() {
			So(err, ShouldBeNil)
			So(conn, ShouldNotBeNil)
		})
		conn, err = GetClientConnection("127", 8183, signedCert, signedKey)
		Convey("Provided a cert and key that is not a CA", func() {
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrCertTrust)
			So(conn, ShouldBeNil)
		})
	})
}
