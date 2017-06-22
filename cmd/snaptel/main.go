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

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/urfave/cli"
)

var (
	gitversion string
	pClient    *client.Client
	timeFormat = time.RFC1123
	err        error
)

type usageError struct {
	s   string
	ctx *cli.Context
}

func (ue usageError) Error() string {
	return fmt.Sprintf("Error: %s \nUsage: %s", ue.s, ue.ctx.Command.Usage)
}

func (ue usageError) help() {
	cli.ShowCommandHelp(ue.ctx, ue.ctx.Command.Name)
}

func newUsageError(s string, ctx *cli.Context) usageError {
	return usageError{s, ctx}
}

func main() {
	app := cli.NewApp()
	app.Name = "snaptel"
	app.Version = gitversion
	app.Usage = "The open telemetry framework"
	app.Flags = []cli.Flag{flURL, flSecure, flAPIVer, flPassword, flConfig, flTimeout}
	app.Commands = append(commands, tribeCommands...)
	sort.Sort(ByCommand(app.Commands))
	app.Before = beforeAction

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		if ue, ok := err.(usageError); ok {
			ue.help()
		}
		os.Exit(1)
	}
}

// Run before every command
func beforeAction(ctx *cli.Context) error {
	username, password := checkForAuth(ctx)
	pClient, err = client.New(ctx.String("url"), ctx.String("api-version"), ctx.Bool("insecure"), client.Timeout(ctx.Duration("timeout")))
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	pClient.Password = password
	pClient.Username = username
	if err = checkTribeCommand(ctx); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// Checks if a tribe command was issued when tribe mode was not
// enabled on the specified snapteld instance.
func checkTribeCommand(ctx *cli.Context) error {
	tribe := false
	for _, a := range os.Args {
		for _, command := range tribeCommands {
			if strings.Contains(a, command.Name) {
				tribe = true
				break
			}
		}
		if tribe {
			break
		}
	}
	if !tribe {
		return nil
	}
	resp := pClient.ListAgreements()
	if resp.Err != nil {
		if resp.Err.Error() == "Invalid credentials" {
			return resp.Err
		}
		return fmt.Errorf("Tribe mode must be enabled in snapteld to use tribe command")
	}
	return nil
}

// Checks for authentication flags and returns a username/password
// from the specified settings
func checkForAuth(ctx *cli.Context) (username, password string) {
	if ctx.Bool("password") {
		username = "snap" // for now since username is unused but needs to exist for basicAuth
		// Prompt for password
		fmt.Print("Password:")
		pass, err := terminal.ReadPassword(0)
		if err != nil {
			password = ""
		} else {
			password = string(pass)
		}
		// Go to next line after password prompt
		fmt.Println()
		return
	}
	//Get config file path in the order:
	if ctx.IsSet("config") {
		cfg := &config{}
		if err := cfg.loadConfig(ctx.String("config")); err != nil {
			fmt.Println(err)
		}
		if cfg.RestAPI.Password != nil {
			// use password declared in config file
			password = *cfg.RestAPI.Password
		} else {
			fmt.Println("Error config password field 'rest-auth-pwd' is empty")
		}
	}
	return
}
