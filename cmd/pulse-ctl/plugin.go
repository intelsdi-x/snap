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
		os.Exit(1)
	}
	err := client.LoadPlugin(ctx.Args().First())
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		os.Exit(1)
	}
}

func listPlugins(ctx *cli.Context) {
	lps, aps, err := client.GetPlugins(ctx.Bool("running"))
	if err != nil {
		fmt.Printf("Error: %v\n", err.Error())
		os.Exit(1)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	if ctx.Bool("running") {
		printFields(w, false, 0, "NAME", "TYPE", "HIT COUNT", "LAST HIT")
		for _, rp := range aps {
			printFields(w, false, 0, rp.Name, rp.TypeName, rp.HitCount, rp.LastHit.Format(time.RFC1123))
		}
	} else {
		printFields(w, false, 0, "NAME", "TYPE", "STATUS", "VERSION", "LOADED TIME")
		for _, lp := range lps {
			printFields(w, false, 0, lp.Name, lp.TypeName, lp.Status, lp.Version, lp.LoadedTimestamp.Format(time.RFC1123))
		}
	}
	w.Flush()
}
