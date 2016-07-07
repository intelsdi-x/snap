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
	"fmt"
	"os"
)

//SNAP_PATH should be set to Snap's build directory

//CheckPluginBuilt checks if PluginName has been built.
func CheckPluginBuilt(SnapPath string, PluginName string) error {
	if SnapPath == "" {
		return fmt.Errorf("SNAP_PATH not set. Cannot test %s plugin.\n", PluginName)
	}
	bPath := fmt.Sprintf("%s/plugin/%s", SnapPath, PluginName)
	if _, err := os.Stat(bPath); os.IsNotExist(err) {
		//the binary has not been built yet
		return fmt.Errorf("Error: $SNAP_PATH/plugin/%s does not exist. Run make to build it.", PluginName)
	}
	return nil
}
