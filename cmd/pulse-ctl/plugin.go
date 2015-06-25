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

	err := client.LoadPlugin(ctx.Args().First())
	if err != nil {
		fmt.Printf("Error loading plugin:\n\t%v\n", err.Error())
		os.Exit(1)
	}
}

func unloadPlugin(ctx *cli.Context) {
	pName := ctx.String("plugin-name")
	pVer := ctx.Int("plugin-version")
	if pName == "" {
		fmt.Println("Must provide plugin name\n")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		os.Exit(1)
	}
	//
	resp := client.UnloadPlugin(pName, pVer)

	fmt.Println(resp)
	// if err != nil {
	// fmt.Printf("Error unloading plugin:\n\t%v\n", err.Error())
	// os.Exit(1)
	// }
	// fmt.Printf("Plugin unloaded successfully (%sv%d)\n", resp.PluginName, resp.PluginVersion)
}

func listPlugins(ctx *cli.Context) {
	plugins := client.GetPlugins(ctx.Bool("running"))
	if plugins.Error != nil {
		fmt.Printf("Error: %v\n", plugins.Error)
		os.Exit(1)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if ctx.Bool("running") {
		printFields(w, false, 0, "NAME", "HIT COUNT", "LAST HIT", "TYPE")
		for _, rp := range plugins.AvailablePlugins {
			printFields(w, false, 0, rp.Name, rp.HitCount, time.Unix(rp.LastHitTimestamp, 0).Format(time.RFC1123), rp.Type)
		}
	} else {
		printFields(w, false, 0, "NAME", "STATUS", "LOADED TIMESTAMP")
		for _, lp := range plugins.LoadedPlugins {
			printFields(w, false, 0, lp.Name, lp.Status, lp.LoadedTimestamp)
		}
	}
	w.Flush()
}
