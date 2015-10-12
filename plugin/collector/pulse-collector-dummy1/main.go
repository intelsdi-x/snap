/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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

package main

import (
	"os"

	// Import the pulse plugin library
	"github.com/intelsdi-x/pulse/control/plugin"
	// Import our collector plugin implementation

	"github.com/intelsdi-x/pulse/plugin/collector/pulse-collector-dummy1/dummy"
)

func main() {
	// Provided:
	//   the definition of the plugin metadata
	//   the implementation satfiying plugin.CollectorPlugin

	// Define metadata about Plugin
	meta := dummy.Meta()
	meta.RPCType = plugin.JSONRPC

	// Start a collector
	plugin.Start(meta, new(dummy.Dummy), os.Args[1])
}
