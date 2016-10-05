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

package psigning

import (
	"io/ioutil"
	"testing"

	"github.com/intelsdi-x/snap/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateSignature(t *testing.T) {
	keyringFile := []string{"pubring.gpg"}
	signedFile := "snap-plugin-collector-mock1"
	signatureFile := signedFile + ".asc"
	unsignedFile := helper.PluginFilePath("snap-plugin-collector-mock2")
	s := SigningManager{}

	signature, _ := ioutil.ReadFile(signatureFile)
	Convey("Valid files and good signature", t, func() {
		err := s.ValidateSignature(keyringFile, signedFile, signature)
		So(err, ShouldBeNil)
	})

	Convey("Valid files and good signature. Multiple keyrings", t, func() {
		keyringFiles := []string{"pubkeys.gpg", "pubring.gpg"}
		err := s.ValidateSignature(keyringFiles, signedFile, signature)
		So(err, ShouldBeNil)
	})

	Convey("Validate unsigned file with signature", t, func() {
		err := s.ValidateSignature(keyringFile, unsignedFile, signature)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Error checking signature")
	})

	Convey("Invalid keyring", t, func() {
		err := s.ValidateSignature([]string{""}, signedFile, signature)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Keyring file (.gpg) not found")
	})

	Convey("Invalid signed file", t, func() {
		err := s.ValidateSignature(keyringFile, "", signature)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Signed file not found")
	})

	Convey("Invalid signature file", t, func() {
		err := s.ValidateSignature(keyringFile, signedFile, nil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Error checking signature")
	})
}
