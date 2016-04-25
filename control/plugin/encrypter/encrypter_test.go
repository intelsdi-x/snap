// +build legacy

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

package encrypter

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEncrypter(t *testing.T) {
	Convey("Encrypter", t, func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		So(err, ShouldBeNil)
		e := New(&key.PublicKey, key)
		symkey, err := GenerateKey()
		So(err, ShouldBeNil)
		e.Key = symkey
		Convey("The constructor works", func() {
			So(e, ShouldHaveSameTypeAs, &Encrypter{})
		})
		Convey("it can encrypt stuff", func() {
			_, err := e.Encrypt(bytes.NewReader([]byte("hello, encrypter")))
			So(err, ShouldBeNil)
		})
		Convey("it can decrypt stuff", func() {
			out, err := e.Encrypt(bytes.NewReader([]byte("hello, encrypter")))
			So(err, ShouldBeNil)
			dec, err := e.Decrypt(bytes.NewReader(out))
			So(err, ShouldBeNil)
			So(string(dec), ShouldEqual, "hello, encrypter")
		})
	})
}
