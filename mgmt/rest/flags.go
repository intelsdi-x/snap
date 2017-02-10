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

package rest

import (
	"fmt"

	"github.com/urfave/cli"
)

var (
	flAPIDisabled = cli.BoolFlag{
		Name:  "disable-api, d",
		Usage: "Disable the agent REST API",
	}
	flAPIAddr = cli.StringFlag{
		Name:   "api-addr, b",
		Usage:  "API Address[:port] to bind to/listen on. Default: empty string => listen on all interfaces",
		EnvVar: "SNAP_ADDR",
	}
	flAPIPort = cli.StringFlag{
		Name:   "api-port, p",
		Usage:  fmt.Sprintf("API port (default: %v)", defaultPort),
		EnvVar: "SNAP_PORT",
	}
	flRestHTTPS = cli.BoolFlag{
		Name:  "rest-https",
		Usage: "start Snap's API as https",
	}
	flRestCert = cli.StringFlag{
		Name:  "rest-cert",
		Usage: "A path to a certificate to use for HTTPS deployment of Snap's REST API",
	}
	flRestKey = cli.StringFlag{
		Name:  "rest-key",
		Usage: "A path to a key file to use for HTTPS deployment of Snap's REST API",
	}
	flRestAuth = cli.BoolFlag{
		Name:  "rest-auth",
		Usage: "Enables Snap's REST API authentication",
	}
	flPProf = cli.BoolFlag{
		Name:  "pprof",
		Usage: "Enables profiling tools",
	}
	flCorsd = cli.StringFlag{
		Name:  "allowed_origins",
		Usage: "Define Cors allowed origins",
	}

	// Flags consumed by snapteld
	Flags = []cli.Flag{flAPIDisabled, flAPIAddr, flAPIPort, flRestHTTPS, flRestCert, flRestKey, flRestAuth, flPProf, flCorsd}
)
