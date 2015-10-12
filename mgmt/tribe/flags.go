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
	"net"
	"os"

	"github.com/codegangsta/cli"
	"github.com/pborman/uuid"
)

var (
	flTribeNodeName = cli.StringFlag{
		Name:   "tribe-node-name",
		Usage:  "Name of this node in tribe cluster (default: hostname)",
		EnvVar: "PULSE_TRIBE_NODE_NAME",
		Value:  getHostname(),
	}

	flTribe = cli.BoolFlag{
		Name:   "tribe",
		Usage:  `Enable tribe mode`,
		EnvVar: "PULSE_TRIBE",
	}

	flTribeSeed = cli.StringFlag{
		Name: "tribe-seed",
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

	flTribeAdvertiseAddr = cli.StringFlag{
		Name:   "tribe-addr",
		Usage:  "Addr tribe gossips over to maintain membership",
		EnvVar: "PULSE_TRIBE_ADDR",
		Value:  getIP(),
	}

	// Flags consumed by pulsed
	Flags = []cli.Flag{flTribeNodeName, flTribe, flTribeSeed, flTribeAdvertiseAddr, flTribeAdvertisePort}
)

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return uuid.New()
	}
	return hostname
}

func getIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.WithField("_block", "getIP").Error(err)
		return "127.0.0.1"
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			logger.WithField("_block", "getIP").Error(err)
			return "127.0.0.1"
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPAddr:
				ip = v.IP
			case *net.IPNet:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}
	return "127.0.0.1"
}
