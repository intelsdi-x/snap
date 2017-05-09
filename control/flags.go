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

import (
	"fmt"

	"github.com/urfave/cli"
)

var (
	flNumberOfPLs = cli.StringFlag{
		Name:   "max-running-plugins, m",
		Usage:  fmt.Sprintf("The maximum number of instances of a loaded plugin to run (default: %v)", defaultMaxRunningPlugins),
		EnvVar: "SNAP_MAX_PLUGINS",
	}
	flPluginLoadTimeout = cli.StringFlag{
		Name:   "plugin-load-timeout",
		Usage:  fmt.Sprintf("The maximum number seconds a plugin can take to load (default: %v)", defaultPluginLoadTimeout),
		EnvVar: "SNAP_PLUGIN_LOAD_TIMEOUT",
	}
	flPluginTrust = cli.StringFlag{
		Name:   "plugin-trust, t",
		Usage:  fmt.Sprintf("0-2 (Disabled, Enabled, Warning; default: %v)", defaultPluginTrust),
		EnvVar: "SNAP_TRUST_LEVEL",
	}

	flAutoDiscover = cli.StringFlag{
		Name:   "auto-discover, a",
		Usage:  "Auto discover paths separated by colons.",
		EnvVar: "SNAP_AUTODISCOVER_PATH",
	}
	flKeyringPaths = cli.StringFlag{
		Name:   "keyring-paths, k",
		Usage:  "Keyring paths for signing verification separated by colons",
		EnvVar: "SNAP_KEYRING_PATHS",
	}
	flCache = cli.StringFlag{
		Name:   "cache-expiration",
		Usage:  fmt.Sprintf("The time limit for which a metric cache entry is valid (default: %v)", defaultCacheExpiration),
		EnvVar: "SNAP_CACHE_EXPIRATION",
	}

	flTLSCert = cli.StringFlag{
		Name:  "tls-cert",
		Usage: "A path to PEM-encoded certificate to use for GRPC TLS channels",
	}
	flTLSKey = cli.StringFlag{
		Name:  "tls-key",
		Usage: "A path to PEM-encoded private key file to use for GRPC TLS channels",
	}
	flCACertPaths = cli.StringFlag{
		Name:  "ca-cert-paths",
		Usage: "A list of paths to CA certificates or their parent directories, separated with OS path separator",
	}

	flControlRpcPort = cli.StringFlag{
		Name:   "control-listen-port",
		Usage:  fmt.Sprintf("Listen port for control RPC server (default: %v)", defaultListenPort),
		EnvVar: "SNAP_CONTROL_LISTEN_PORT",
	}

	flControlRpcAddr = cli.StringFlag{
		Name:   "control-listen-addr",
		Usage:  "Listen address for control RPC server",
		EnvVar: "SNAP_CONTROL_LISTEN_ADDR",
	}

	flTempDirPath = cli.StringFlag{
		Name:   "temp_dir_path",
		Usage:  fmt.Sprintf("Temporary path for loading plugins (default: %v)", defaultTempDirPath),
		EnvVar: "SNAP_TEMP_DIR_PATH",
	}

	Flags = []cli.Flag{flNumberOfPLs, flPluginLoadTimeout, flAutoDiscover, flPluginTrust, flKeyringPaths, flCache, flControlRpcPort, flControlRpcAddr, flTempDirPath, flTLSCert, flTLSKey, flCACertPaths}
)
