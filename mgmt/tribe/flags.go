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

package tribe

import (
	"math/rand"
	"os"

	"github.com/codegangsta/cli"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	flTribeNodeName = cli.StringFlag{
		Name:   "tribe-node-name",
		Usage:  "Name of this node in tribe cluster (default: hostname)",
		EnvVar: "PULSE_TRIBE_NODE_NAME",
		Value:  getHostname(),
	}

	flTribeSeed = cli.StringFlag{
		Name: "tribe",
		Usage: `IP or resolvable hostname of a seed node to join.
	The default empty value assumes this is the first node in a cluster.`,
		EnvVar: "PULSE_TRIBE_SEED",
		Value:  "",
	}

	flTribeAdvertisePort = cli.IntFlag{
		Name:   "tribe-port",
		Usage:  "Port tribe gossips over to maintain membership",
		EnvVar: "PULSE_TRIBE_PORT",
		Value:  6000,
	}

	Flags = []cli.Flag{flTribeNodeName, flTribeSeed, flTribeAdvertisePort}
)

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return randHostname(8)
	}
	return hostname
}

func randHostname(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
