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

package core

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/intelsdi-x/snap/plugin/helper"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginName    = "snap-plugin-collector-mock1"
	SnapPath      = helper.BuildPath
	PluginPath    = helper.PluginFilePath(PluginName)
	SignatureFile = path.Join(SnapPath, "../pkg/psigning", "snap-plugin-collector-mock1.asc")
	TempPath      = os.TempDir()
)

func TestRequestedPlugin(t *testing.T) {
	// Creating a plugin request

	Convey("Creating a plugin request from a valid path", t, func() {
		rp, err := NewRequestedPlugin(PluginPath, TempPath, nil)
		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)

			Convey("Should return a RequestedPlugin type", func() {
				So(rp, ShouldHaveSameTypeAs, &RequestedPlugin{})
			})

			Convey("Should set the path to the plugin", func() {
				So(rp.Path(), ShouldContainSubstring, TempPath)
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
			// Set a signature for the plugin
			rp.SetSignature([]byte{00, 00, 00})
			Convey("A signature for the plugin can be added to a plugin request", func() {
				Convey("So signature should not be nil", func() {
					So(rp.Signature(), ShouldNotBeNil)
				})
				Convey("And signature should equal what we set it to", func() {
					So(rp.Signature(), ShouldResemble, []byte{00, 00, 00})
				})
			})

		})
	})

	Convey("A signature file can be read in for a plugin request", t, func() {
		rp1, err1 := NewRequestedPlugin(PluginPath, TempPath, nil)
		So(err1, ShouldBeNil)
		err1 = rp1.ReadSignatureFile(SignatureFile)

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
	_, err2 := NewRequestedPlugin(PluginPath+"foo", TempPath, nil)
	Convey("An error should be generated when creating a plugin request with non-existent path", t, func() {
		Convey("So error should not be nil", func() {
			So(err2, ShouldNotBeNil)
		})
	})
	// Create a plugin request and try to add a signature from an non-existent signature file

	Convey("When passing in a non-existent signature file", t, func() {
		rp3, err3 := NewRequestedPlugin(PluginPath, TempPath, nil)
		So(err3, ShouldBeNil)
		err3 = rp3.ReadSignatureFile(SignatureFile + "foo")

		Convey("signature should still be nil", func() {
			So(rp3.Signature(), ShouldBeNil)
		})
		Convey("and error should be returned", func() {
			So(err3, ShouldNotBeNil)
		})
	})
}
