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

package control

import "github.com/codegangsta/cli"

var (
	flNumberOfPLs = cli.IntFlag{
		Name:   "max-running-plugins, m",
		Usage:  "The maximum number of instances of a loaded plugin to run",
		EnvVar: "SNAP_MAX_PLUGINS",
	}

	flAutoDiscover = cli.StringFlag{
		Name:   "auto-discover, a",
		Usage:  "Auto discover paths separated by colons.",
		EnvVar: "SNAP_AUTODISCOVER_PATH",
	}

	flPluginTrust = cli.IntFlag{
		Name:   "plugin-trust, t",
		Usage:  "0-2 (Disabled, Enabled, Warning)",
		EnvVar: "SNAP_TRUST_LEVEL",
	}

	flKeyringPaths = cli.StringFlag{
		Name:   "keyring-paths, k",
		Usage:  "Keyring paths for signing verification separated by colons",
		EnvVar: "SNAP_KEYRING_PATHS",
	}

	flCache = cli.DurationFlag{
		Name:   "cache-expiration",
		Usage:  "The time limit for which a metric cache entry is valid",
		EnvVar: "SNAP_CACHE_EXPIRATION",
	}
	flControlRpcPort = cli.IntFlag{
		Name:   "control-listen-port",
		Usage:  "Listen port for control RPC server",
		EnvVar: "SNAP_CONTROL_LISTEN_PORT",
	}

	flControlRpcAddr = cli.StringFlag{
		Name:   "control-listen-addr",
		Usage:  "Listen address for control RPC server",
		EnvVar: "SNAP_CONTROL_LISTEN_ADDR",
	}

	Flags = []cli.Flag{flNumberOfPLs, flAutoDiscover, flPluginTrust, flKeyringPaths, flCache, flControlRpcPort, flControlRpcAddr}
)
