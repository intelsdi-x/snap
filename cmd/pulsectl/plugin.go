package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
)

func loadPlugin(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		fmt.Print("Incorrect usage\n")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		os.Exit(1)
	}
	r := pClient.LoadPlugin(ctx.Args().First())
	if r.Err != nil {
		fmt.Printf("Error loading plugin:\n%v\n", r.Err.Error())
		os.Exit(1)
	}
	for _, p := range r.LoadedPlugins {
		fmt.Println("Plugin loaded")
		fmt.Printf("Name: %s\n", p.Name)
		fmt.Printf("Version: %d\n", p.Version)
		fmt.Printf("Type: %s\n", p.Type)
		fmt.Printf("Loaded Time: %s\n\n", p.LoadedTime().Format(timeFormat))
	}
}

func unloadPlugin(ctx *cli.Context) {
	pName := ctx.String("plugin-name")
	pVer := ctx.Int("plugin-version")
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

	r := pClient.UnloadPlugin(pName, pVer)
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
		printFields(w, false, 0, "NAME", "VERSION", "TYPE", "STATUS", "LOADED TIME")
		for _, lp := range plugins.LoadedPlugins {
			printFields(w, false, 0, lp.Name, lp.Version, lp.Type, lp.Status, lp.LoadedTime().Format(timeFormat))
		}
	}
	w.Flush()
}
