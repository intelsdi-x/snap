package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

func listMetrics(ctx *cli.Context) {
	mts, err := client.GetMetricCatalog()
	if err != nil {
		fmt.Printf("error getting metric catalog: %v", err)
		os.Exit(1)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Namespace\tVersion")
	for _, mt := range mts {
		fmt.Fprintf(w, "%v\t%v\n", mt.Namespace, mt.Version)
	}
	w.Flush()
}
