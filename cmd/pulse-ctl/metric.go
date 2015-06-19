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
	printFields(w, false, 0, "NAMESPACE", "VERSION")
	for _, mt := range mts {
		printFields(w, false, 0, mt.Namespace, mt.Version)
	}
	w.Flush()
}
