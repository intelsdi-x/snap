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
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
)

func loadPlugin(ctx *cli.Context) {
	pAsc := ctx.String("plugin-asc")
	var paths []string
	if len(ctx.Args()) != 1 {
		fmt.Println("Incorrect usage:")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		os.Exit(1)
	}
	paths = append(paths, ctx.Args().First())
	if pAsc != "" {
		if !strings.Contains(pAsc, ".asc") {
			fmt.Println("Must be a .asc file for the -a flag")
			cli.ShowCommandHelp(ctx, ctx.Command.Name)
			os.Exit(1)
		}
		paths = append(paths, pAsc)
	}
	r := pClient.LoadPlugin(paths)
	if r.Err != nil {
		if r.Err.Fields()["error"] != nil {
			fmt.Printf("Error loading plugin:\n%v\n%v\n", r.Err.Error(), r.Err.Fields()["error"])
		} else {
			fmt.Printf("Error loading plugin:\n%v\n", r.Err.Error())
		}
		os.Exit(1)
	}
	for _, p := range r.LoadedPlugins {
		fmt.Println("Plugin loaded")
		fmt.Printf("Name: %s\n", p.Name)
		fmt.Printf("Version: %d\n", p.Version)
		fmt.Printf("Type: %s\n", p.Type)
		fmt.Printf("Signed: %v\n", p.Signed)
		fmt.Printf("Loaded Time: %s\n\n", p.LoadedTime().Format(timeFormat))
	}
}

func unloadPlugin(ctx *cli.Context) {
	pDetails := filepath.SplitList(ctx.Args().First())
	var pType string
	var pName string
	var pVer int
	var err error

	if len(pDetails) == 3 {
		pType = pDetails[0]
		pName = pDetails[1]
		pVer, err = strconv.Atoi(pDetails[2])
		if err != nil {
			fmt.Println("Can't convert version string to integer")
			cli.ShowCommandHelp(ctx, ctx.Command.Name)
			os.Exit(1)
		}
	} else {
		pType = ctx.String("plugin-type")
		pName = ctx.String("plugin-name")
		pVer = ctx.Int("plugin-version")
	}
	if pType == "" {
		fmt.Println("Must provide plugin type")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		os.Exit(1)
	}
	if pName == "" {
		fmt.Println("Must provide plugin name")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		os.Exit(1)
	}
	if pVer < 1 {
		fmt.Println("Must provide plugin version")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		os.Exit(1)
	}

	r := pClient.UnloadPlugin(pType, pName, pVer)
	if r.Err != nil {
		fmt.Printf("Error unloading plugin:\n%v\n", r.Err.Error())
		os.Exit(1)
	}

	fmt.Println("Plugin unloaded")
	fmt.Printf("Name: %s\n", r.Name)
	fmt.Printf("Version: %d\n", r.Version)
	fmt.Printf("Type: %s\n", r.Type)
}

func listPlugins(ctx *cli.Context) {
	plugins := pClient.GetPlugins(ctx.Bool("running"))
	if plugins.Err != nil {
		fmt.Printf("Error: %v\n", plugins.Err)
		os.Exit(1)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if ctx.Bool("running") {
		printFields(w, false, 0, "NAME", "HIT COUNT", "LAST HIT", "TYPE")
		for _, rp := range plugins.AvailablePlugins {
			printFields(w, false, 0, rp.Name, rp.HitCount, time.Unix(rp.LastHitTimestamp, 0).Format(timeFormat), rp.Type)
		}
	} else {
		printFields(w, false, 0, "NAME", "VERSION", "TYPE", "SIGNED", "STATUS", "LOADED TIME")
		for _, lp := range plugins.LoadedPlugins {
			printFields(w, false, 0, lp.Name, lp.Version, lp.Type, lp.Signed, lp.Status, lp.LoadedTime().Format(timeFormat))
		}
	}
	w.Flush()
}
