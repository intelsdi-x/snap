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
	"path"
	"runtime"
)

var (
	BuildPath = os.ExpandEnv(os.Getenv("SNAP_PATH"))
)

func PluginFileCheck(name string) error {
	if BuildPath == "" {
		return fmt.Errorf("$SNAP_PATH environment variable not set. Cannot test plugins.\n")
	}

	fpath := PluginFilePath(name)
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		//the binary has not been built yet
		return fmt.Errorf("Error: %s does not exist in %s. Run make to build it.", name, fpath)
	}
	return nil
}

func PluginPath() string {
	var arch string
	if runtime.GOARCH == "amd64" {
		arch = "x86_64"
	} else {
		arch = runtime.GOARCH
	}

	fpath := path.Join(BuildPath, runtime.GOOS, arch, "plugins")
	return fpath
}

func PluginFilePath(name string) string {
	fpath := path.Join(PluginPath(), name)
	return fpath
}
