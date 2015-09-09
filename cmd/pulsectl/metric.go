package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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
		NAMESPACE               VERSION
		/intel/dummy/foo        1,2
		/intel/dummy/bar        1
	*/
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	metsByVer := make(map[string][]string)
	for _, mt := range mts.Catalog {
		metsByVer[mt.Namespace] = append(metsByVer[mt.Namespace], strconv.Itoa(mt.Version))
	}
	printFields(w, false, 0, "NAMESPACE", "VERSIONS")
	for ns, vers := range metsByVer {
		printFields(w, false, 0, ns, strings.Join(vers, ","))
	}
	w.Flush()
	return
}

func getMetric(ctx *cli.Context) {
	ns := ctx.String("metric-namespace")
	ver := ctx.Int("metric-version")
	if ns == "" {
		fmt.Println("namespace is required")
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		return
	}
	if ver == 0 {
		ver = -1
	}
	metrics := pClient.FetchMetrics(ns, ver)
	if metrics.Err != nil {
		fmt.Println(metrics.Err)
		return
	}
	metric := metrics.Catalog[0]

	/*
		NAMESPACE                VERSION         LAST ADVERTISED TIME
		/intel/dummy/foo         2               Wed, 09 Sep 2015 10:01:04 PDT

		  Rules for collecting /intel/dummy/foo:

		     NAME        TYPE            DEFAULT         REQUIRED
		     name        string          bob             false
		     password    string                          true
	*/

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0, "NAMESPACE", "VERSION", "LAST ADVERTISED TIME")
	printFields(w, false, 0, metric.Namespace, metric.Version, time.Unix(metric.LastAdvertisedTimestamp, 0).Format(time.RFC1123))
	w.Flush()
	fmt.Printf("\n  Rules for collecting %s:\n\n", metric.Namespace)
	printFields(w, true, 4, "NAME", "TYPE", "DEFAULT", "REQUIRED")
	for _, rule := range metric.Policy {
		defMap, ok := rule.Default.(map[string]interface{})
		if ok {
			def := defMap["Value"]
			printFields(w, true, 4, rule.Name, rule.Type, def, rule.Required)
		} else {
			printFields(w, true, 4, rule.Name, rule.Type, "", rule.Required)
		}
	}
	w.Flush()
}
