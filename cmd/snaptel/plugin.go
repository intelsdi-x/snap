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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli"
)

func loadPlugin(ctx *cli.Context) error {
	pAsc := ctx.String("plugin-asc")
	pCert := ctx.String("plugin-cert")
	pKey := ctx.String("plugin-key")
	var paths []string
	if len(ctx.Args()) >3 {
		return newUsageError("Incorrect usage:", ctx)
	}
	paths = append(paths, ctx.Args().First())
	if pAsc != "" {
		if !strings.Contains(pAsc, ".asc") {
			return newUsageError("Must be a .asc file for the -a flag", ctx)
		}
		paths = append(paths, pAsc)
	}
	if pCert != "" {
		tmpFile, err := ioutil.TempFile("", "crt.")
		if err != nil {
			return fmt.Errorf("Error processing plugin certificate - unable to create link:\n%v", err.Error())
		}
		_, err = tmpFile.WriteString(pCert)
		if err != nil {
			return fmt.Errorf("Error processing plugin certificate - unable to write link:\n%v", err.Error())
		}
		paths = append(paths, tmpFile.Name())
	}
	if pKey != "" {
		tmpFile, err := ioutil.TempFile("", "key.")
		if err != nil {
			return fmt.Errorf("Error processing plugin key - unable to create link:\n%v", err.Error())
		}
		_, err = tmpFile.WriteString(pKey)
		if err != nil {
			return fmt.Errorf("Error processing plugin key - unable to write link:\n%v", err.Error())
		}
		paths = append(paths, tmpFile.Name())
	}
	r := pClient.LoadPlugin(paths)
	if r.Err != nil {
		if r.Err.Fields()["error"] != nil {
			return fmt.Errorf("Error loading plugin:\n%v\n%v\n", r.Err.Error(), r.Err.Fields()["error"])
		}
		return fmt.Errorf("Error loading plugin:\n%v\n", r.Err.Error())
	}
	for _, p := range r.LoadedPlugins {
		fmt.Println("Plugin loaded")
		fmt.Printf("Name: %s\n", p.Name)
		fmt.Printf("Version: %d\n", p.Version)
		fmt.Printf("Type: %s\n", p.Type)
		fmt.Printf("Signed: %v\n", p.Signed)
		fmt.Printf("Loaded Time: %s\n\n", p.LoadedTime().Format(timeFormat))
	}

	return nil
}

func unloadPlugin(ctx *cli.Context) error {
	pType := ctx.Args().Get(0)
	pName := ctx.Args().Get(1)
	pVer, err := strconv.Atoi(ctx.Args().Get(2))

	if pType == "" {
		return newUsageError("Must provide plugin type", ctx)
	}
	if pName == "" {
		return newUsageError("Must provide plugin name", ctx)
	}
	if err != nil {
		return newUsageError("Can't convert version string to integer", ctx)
	}
	if pVer < 1 {
		return newUsageError("Must provide plugin version", ctx)
	}

	r := pClient.UnloadPlugin(pType, pName, pVer)
	if r.Err != nil {
		return fmt.Errorf("Error unloading plugin:\n%v\n", r.Err.Error())
	}

	fmt.Println("Plugin unloaded")
	fmt.Printf("Name: %s\n", r.Name)
	fmt.Printf("Version: %d\n", r.Version)
	fmt.Printf("Type: %s\n", r.Type)

	return nil
}

func swapPlugins(ctx *cli.Context) error {
	// plugin to load
	pAsc := ctx.String("plugin-asc")
	var paths []string
	if len(ctx.Args()) < 1 || len(ctx.Args()) > 2 {
		return newUsageError("Incorrect usage:", ctx)
	}
	paths = append(paths, ctx.Args().First())
	if pAsc != "" {
		if !strings.Contains(pAsc, ".asc") {
			return newUsageError("Must be a .asc file for the -a flag", ctx)
		}
		paths = append(paths, pAsc)
	}

	// plugin to unload
	var pDetails []string
	var pType, pName string
	var pVer int
	var err error

	if len(ctx.Args()) == 2 {
		pDetails = filepath.SplitList(ctx.Args()[1])
		if len(pDetails) == 3 {
			pType = pDetails[0]
			pName = pDetails[1]
			pVer, err = strconv.Atoi(pDetails[2])
			if err != nil {
				return newUsageError("Can't convert version string to integer", ctx)
			}
		} else {
			return newUsageError("Missing type, name, or version", ctx)
		}
	} else {
		pType = ctx.String("plugin-type")
		pName = ctx.String("plugin-name")
		pVer = ctx.Int("plugin-version")
	}
	if pType == "" {
		return newUsageError("Must provide plugin type", ctx)
	}
	if pName == "" {
		return newUsageError("Must provide plugin name", ctx)
	}
	if pVer < 1 {
		return newUsageError("Must provide plugin version", ctx)
	}

	r := pClient.SwapPlugin(paths, pType, pName, pVer)
	if r.Err != nil {
		return fmt.Errorf("Error swapping plugins:\n%v\n", r.Err.Error())
	}

	fmt.Println("Plugin loaded")
	fmt.Printf("Name: %s\n", r.LoadedPlugin.Name)
	fmt.Printf("Version: %d\n", r.LoadedPlugin.Version)
	fmt.Printf("Type: %s\n", r.LoadedPlugin.Type)
	fmt.Printf("Signed: %v\n", r.LoadedPlugin.Signed)
	fmt.Printf("Loaded Time: %s\n\n", r.LoadedPlugin.LoadedTime().Format(timeFormat))

	fmt.Println("\nPlugin unloaded")
	fmt.Printf("Name: %s\n", r.UnloadedPlugin.Name)
	fmt.Printf("Version: %d\n", r.UnloadedPlugin.Version)
	fmt.Printf("Type: %s\n", r.UnloadedPlugin.Type)

	return nil
}

func listPlugins(ctx *cli.Context) error {
	plugins := pClient.GetPlugins(ctx.Bool("running"))
	if plugins.Err != nil {
		return fmt.Errorf("Error: %v\n", plugins.Err)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if ctx.Bool("running") {
		if len(plugins.AvailablePlugins) == 0 {
			fmt.Println("No running plugins found. Have you started a task?")
			return nil
		}
		printFields(w, false, 0, "NAME", "HIT COUNT", "LAST HIT", "TYPE", "PPROF PORT")
		for _, rp := range plugins.AvailablePlugins {
			printFields(w, false, 0, rp.Name, rp.HitCount, time.Unix(rp.LastHitTimestamp, 0).Format(timeFormat), rp.Type, rp.PprofPort)
		}
	} else {
		if len(plugins.LoadedPlugins) == 0 {
			fmt.Println("No plugins found. Have you loaded a plugin?")
			return nil
		}
		printFields(w, false, 0, "NAME", "VERSION", "TYPE", "SIGNED", "STATUS", "LOADED TIME")
		for _, lp := range plugins.LoadedPlugins {
			printFields(w, false, 0, lp.Name, lp.Version, lp.Type, lp.Signed, lp.Status, lp.LoadedTime().Format(timeFormat))
		}
	}
	w.Flush()

	return nil
}
