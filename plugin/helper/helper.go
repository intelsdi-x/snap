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

// Package helper The test helper for testing plugins
package helper

import (
	// "fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	buildScript = "/scripts/build.sh"
)

// BuildPlugin attempts to make the plugins before each test.
func BuildPlugin(pluginType, pluginName string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	bPath := strings.Replace(wd, path.Join("/", "plugin", pluginType, pluginName), buildScript, 1)
	sPath := strings.Replace(wd, path.Join("/", "plugin", pluginType, pluginName), "", 1)

	// fmt.Println(bPath, sPath, pluginType, pluginName)
	c := exec.Command(bPath, sPath, pluginType, pluginName)

	_, e := c.Output()
	// fmt.Println(string(o))
	return e
}
