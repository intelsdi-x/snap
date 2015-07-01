package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

func listMetrics(ctx *cli.Context) {
	mts := pClient.GetMetricCatalog()
	if mts.Err != nil {
		fmt.Printf("Error getting metric catalog: %v", mts.Err)
		os.Exit(1)
	}

	/*
		NAMESPACE			VERSION
		/intel/dummy/foo	1,2
		/intel/dummy/bar	1
	*/
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0, "NAMESPACE", "VERSION")
	for _, mt := range mts.Catalog {
		v := make([]string, 0)
		for k, _ := range mt.Versions {
			v = append(v, k)
		}
		printFields(w, false, 0, mt.Namespace, strings.Join(v, ","))
	}
	w.Flush()
}
