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
		fmt.Fprintln(w, "Name\tHit count\tLast Hit\tType")
		for _, rp := range aps {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", rp.Name, rp.HitCount, rp.LastHit.Format(time.RFC1123), rp.TypeName)
		}
	} else {
		fmt.Fprintln(w, "Name\tStatus\tVersion\tLoaded Timestamp")
		for _, lp := range lps {
			fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", lp.Name, lp.Status, lp.Version, time.Unix(lp.LoadedTimestamp, 0))
		}
	}

	w.Flush()
}
