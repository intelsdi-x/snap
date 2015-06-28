package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

func listMetrics(ctx *cli.Context) {
	mts := pClient.GetMetricCatalog()
	if mts.Err != nil {
		fmt.Printf("error getting metric catalog: %v", mts.Err)
		os.Exit(1)
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0, "NAMESPACE", "VERSION")
	for _, mt := range mts.Catalog {
		printFields(w, false, 0, mt.Namespace)
	}
	w.Flush()
}
