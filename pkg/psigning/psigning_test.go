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

package psigning

import (
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateSignature(t *testing.T) {
	keyringFile := "pubring.gpg"
	signedFile := "pulse-collector-dummy1"
	signatureFile := signedFile + ".asc"
	pulsePath := os.Getenv("PULSE_PATH")
	unsignedFile := path.Join(pulsePath, "plugin", "pulse-collector-dummy2")

	s := SigningManager{}

	Convey("Valid files and good signature", t, func() {
		err := s.ValidateSignature(keyringFile, signedFile, signatureFile)
		So(err, ShouldBeNil)
	})

	Convey("Valid files and bad signature", t, func() {
		err := s.ValidateSignature(keyringFile, unsignedFile, signatureFile)
		So(err.Error(), ShouldResemble, "Error checking signature")
	})

	Convey("Invalid keyring", t, func() {
		err := s.ValidateSignature("", signedFile, signatureFile)
		So(err.Error(), ShouldResemble, "Keyring file (.gpg) not found")
	})

	Convey("Invalid signed file", t, func() {
		err := s.ValidateSignature(keyringFile, "", signatureFile)
		So(err.Error(), ShouldResemble, "Signed file not found")
	})

	Convey("Invalid signature file", t, func() {
		err := s.ValidateSignature(keyringFile, signedFile, "")
		So(err.Error(), ShouldResemble, "Signature file (.asc) not found")
	})
}
