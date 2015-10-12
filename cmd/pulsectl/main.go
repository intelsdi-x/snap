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
	"time"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/pulse/mgmt/rest/client"
)

var (
	gitversion string
	pClient    *client.Client
	timeFormat = time.RFC1123
)

func main() {

	app := cli.NewApp()
	app.Name = "pulsectl"
	app.Version = gitversion
	app.Usage = "A powerful telemetry agent framework"
	app.Flags = []cli.Flag{flURL, flSecure, flAPIVer}
	app.Commands = commands

	app.Before = func(c *cli.Context) error {
		if pClient == nil {
			pClient = client.New(c.GlobalString("url"), c.GlobalString("api-version"), c.GlobalBool("insecure"))
		}
		return nil
	}

	app.Run(os.Args)
}
