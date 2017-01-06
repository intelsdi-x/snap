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

package tribe

import (
	"fmt"

	"github.com/urfave/cli"
)

var (
	flTribeNodeName = cli.StringFlag{
		Name:   "tribe-node-name",
		Usage:  "Name of this node in tribe cluster (default: hostname)",
		EnvVar: "SNAP_TRIBE_NODE_NAME",
	}

	flTribe = cli.BoolFlag{
		Name:   "tribe",
		Usage:  `Enable tribe mode`,
		EnvVar: "SNAP_TRIBE",
	}

	flTribeSeed = cli.StringFlag{
		Name:   "tribe-seed",
		Usage:  "IP (or hostname) and port of a node to join (e.g. 127.0.0.1:6000)",
		EnvVar: "SNAP_TRIBE_SEED",
	}

	flTribeAdvertisePort = cli.StringFlag{
		Name:   "tribe-port",
		Usage:  fmt.Sprintf("Port tribe gossips over to maintain membership (default: %v)", defaultBindPort),
		EnvVar: "SNAP_TRIBE_PORT",
	}

	flTribeAdvertiseAddr = cli.StringFlag{
		Name:   "tribe-addr",
		Usage:  "Addr tribe gossips over to maintain membership",
		EnvVar: "SNAP_TRIBE_ADDR",
	}

	// Flags consumed by snapteld
	Flags = []cli.Flag{flTribeNodeName, flTribe, flTribeSeed, flTribeAdvertiseAddr, flTribeAdvertisePort}
)
