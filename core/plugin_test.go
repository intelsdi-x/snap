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

package core

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName    = "pulse-collector-mock1"
	PulsePath     = os.Getenv("PULSE_PATH")
	PluginPath    = path.Join(PulsePath, "plugin", PluginName)
	SignatureFile = path.Join(PulsePath, "../pkg/psigning", "pulse-collector-mock1.asc")
)

func TestRequestedPlugin(t *testing.T) {
	// Creating a plugin request
	rp, err := NewRequestedPlugin(PluginPath)
	Convey("Creating a plugin request from a valid path", t, func() {
		Convey("Should return a RequestedPlugin type", func() {
			So(rp, ShouldHaveSameTypeAs, &RequestedPlugin{})
		})
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)
		})
		Convey("Should set the path to the plugin", func() {
			So(rp.Path(), ShouldEqual, PluginPath)
		})
		Convey("Should generate a checksum for the plugin", func() {
			So(rp.CheckSum(), ShouldNotBeNil)
			Convey("And checksum should match manually generated checksum", func() {
				b, _ := ioutil.ReadFile(PluginPath)
				So(rp.CheckSum(), ShouldResemble, sha256.Sum256(b))
			})
		})
		Convey("And plugin signature should initially be nil", func() {
			So(rp.Signature(), ShouldBeNil)
		})
	})
	// Set a signature for the plugin
	rp.SetSignature([]byte{00, 00, 00})
	Convey("A signature for the plugin can be added to a plugin request", t, func() {
		Convey("So signature should not be nil", func() {
			So(rp.Signature(), ShouldNotBeNil)
		})
		Convey("And signature should equal what we set it to", func() {
			So(rp.Signature(), ShouldResemble, []byte{00, 00, 00})
		})
	})
	// Create a plugin request and read a signature file for the plugin
	rp1, _ := NewRequestedPlugin(PluginPath)
	err1 := rp1.ReadSignatureFile(SignatureFile)
	Convey("A signature file can be read in for a plugin request", t, func() {
		Convey("Should not receive an error reading signature file", func() {
			So(err1, ShouldBeNil)
		})
		Convey("So signature should not be nil", func() {
			So(rp1.Signature(), ShouldNotBeNil)
		})
		Convey("So signature should match signature from file", func() {
			b, _ := ioutil.ReadFile(SignatureFile)
			So(rp1.Signature(), ShouldResemble, b)
		})
	})
	// Try to create a plugin request from a bad path to a plugin
	_, err2 := NewRequestedPlugin(PluginPath + "foo")
	Convey("An error should be generated when creating a plugin request with non-existant path", t, func() {
		Convey("So error should not be nil", func() {
			So(err2, ShouldNotBeNil)
		})
	})
	// Create a plugin request and try to add a signature from an non-existant signature file
	rp3, _ := NewRequestedPlugin(PluginPath)
	err3 := rp3.ReadSignatureFile(SignatureFile + "foo")
	Convey("When passing in a non-existant signature file", t, func() {
		Convey("signature should still be nil", func() {
			So(rp3.Signature(), ShouldBeNil)
		})
		Convey("and error should be returned", func() {
			So(err3, ShouldNotBeNil)
		})
	})
}
