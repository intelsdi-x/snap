package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/codegangsta/cli"
)

func listMetrics(ctx *cli.Context) {
	ns := ctx.String("metric-namespace")
	ver := ctx.Int("metric-version")
	if ns != "" {
		//if the user doesn't provide '/*' we fix it
		if ns[len(ns)-2:] != "/*" {
			if ns[len(ns)-1:] == "/" {
				ns = ns + "*"
			} else {
				ns = ns + "/*"
			}
		}
	} else {
		ns = "/*"
	}
	mts := pClient.FetchMetrics(ns, ver)
	if mts.Err != nil {
		fmt.Printf("Error getting metrics: %v\n", mts.Err)
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
		printFields(w, false, 0, mt.Namespace, strings.Join(sortVersions(v), ","))
	}
	w.Flush()
}

func sortVersions(vers []string) []string {
	ivers := make([]int, len(vers))
	svers := make([]string, len(vers))
	var err error
	for i, v := range vers {
		if ivers[i], err = strconv.Atoi(v); err != nil {
			fmt.Printf("Error metric version err: %v", err)
			os.Exit(1)
		}
	}
	sort.Ints(ivers)
	for i, v := range ivers {
		svers[i] = fmt.Sprintf("%d", v)
	}
	return svers
}
