/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
	"github.com/urfave/cli"

	"github.com/intelsdi-x/snap/pkg/stringutils"
)

func listMetrics(ctx *cli.Context) error {
	ns := ctx.String("metric-namespace")
	ver := ctx.Int("metric-version")
	verbose := ctx.Bool("verbose")
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
		return fmt.Errorf("Error getting metrics: %v\n", mts.Err)
	}
	if mts.Len() == 0 {
		fmt.Println("No metrics found. Have you loaded any collectors yet?")
		return nil
	}
	/*
		NAMESPACE               VERSION
		/intel/mock/foo         1,2
		/intel/mock/bar         1
	*/
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)

	if verbose {

		//	NAMESPACE                VERSION         UNIT          DESCRIPTION
		//	/intel/mock/foo          1
		//      /intel/mock/foo          2               mock unit     mock description
		//      /intel/mock/[host]/baz   2               mock unit     mock description

		printFields(w, false, 0, "NAMESPACE", "VERSION", "UNIT", "DESCRIPTION")
		for _, mt := range mts.Catalog {
			namespace := getNamespace(mt)
			printFields(w, false, 0, namespace, mt.Version, mt.Unit, mt.Description)
		}
		w.Flush()
		return nil
	}
	metsByVer := make(map[string][]string)
	for _, mt := range mts.Catalog {
		metsByVer[mt.Namespace] = append(metsByVer[mt.Namespace], strconv.Itoa(mt.Version))
	}
	//make list in alphabetical order
	var key []string
	for k := range metsByVer {
		key = append(key, k)
	}
	sort.Strings(key)

	printFields(w, false, 0, "NAMESPACE", "VERSIONS")
	for _, ns := range key {
		printFields(w, false, 0, ns, strings.Join(metsByVer[ns], ","))
	}
	w.Flush()
	return nil
}

func printMetric(metric *client.GetMetricResult, idx int) error {
	if metric.Err != nil {
		return fmt.Errorf("%v", metric.Err)
	}

	/*
		NAMESPACE                VERSION         LAST ADVERTISED TIME
		/intel/mock/foo          2               Wed, 09 Sep 2015 10:01:04 PDT

		  Rules for collecting /intel/mock/foo:

		     NAME        TYPE            DEFAULT         REQUIRED     MINIMUM   MAXIMUM
		     name        string          bob             false
		     password    string                          true
		     portRange   int                             false        9000      10000
	*/

	namespace := getNamespace(metric.Metric)

	if idx > 0 {
		fmt.Printf("\n")
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0, "NAMESPACE", "VERSION", "UNIT", "LAST ADVERTISED TIME", "DESCRIPTION")
	printFields(w, false, 0, namespace, metric.Metric.Version, metric.Metric.Unit, time.Unix(metric.Metric.LastAdvertisedTimestamp, 0).Format(time.RFC1123), metric.Metric.Description)
	w.Flush()
	if metric.Metric.Dynamic {

		//	NAMESPACE                VERSION     UNIT        LAST ADVERTISED TIME            DESCRIPTION
		//	/intel/mock/[host]/baz   2           mock unit   Wed, 09 Sep 2015 10:01:04 PDT   mock description
		//
		//	  Dynamic elements of namespace: /intel/mock/[host]/baz
		//
		//           NAME        DESCRIPTION
		//           host        name of the host
		//
		//	  Rules for collecting /intel/mock/[host]/baz:
		//
		//	     NAME        TYPE            DEFAULT         REQUIRED     MINIMUM   MAXIMUM

		fmt.Printf("\n  Dynamic elements of namespace: %s\n\n", namespace)
		printFields(w, true, 6, "NAME", "DESCRIPTION")
		for _, v := range metric.Metric.DynamicElements {
			printFields(w, true, 6, v.Name, v.Description)
		}
		w.Flush()
	}
	fmt.Printf("\n  Rules for collecting %s:\n\n", namespace)
	printFields(w, true, 6, "NAME", "TYPE", "DEFAULT", "REQUIRED", "MINIMUM", "MAXIMUM")
	for _, rule := range metric.Metric.Policy {
		printFields(w, true, 6, rule.Name, rule.Type, rule.Default, rule.Required, rule.Minimum, rule.Maximum)
	}
	w.Flush()
	return nil
}

func getMetric(ctx *cli.Context) error {
	if !ctx.IsSet("metric-namespace") {
		return newUsageError("Must provide metric namespace", ctx)
	}
	ns := ctx.String("metric-namespace")
	ver := ctx.Int("metric-version")
	metric := pClient.GetMetric(ns, ver)
	switch mtype := metric.(type) {
	case []*client.GetMetricResult:
		// Multiple metrics
		var merr error
		for i, m := range metric.([]*client.GetMetricResult) {
			err := printMetric(m, i)
			if err != nil {
				merr = err
			}
		}
		if merr != nil {
			return merr
		}
	case *client.GetMetricResult:
		// Single metric
		err := printMetric(metric.(*client.GetMetricResult), 0)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unexpected response type %T\n", mtype)
	}
	return nil
}

func getNamespace(mt *rbody.Metric) string {
	ns := mt.Namespace
	if mt.Dynamic {
		fc := stringutils.GetFirstChar(ns)
		slice := strings.Split(ns, fc)
		for _, v := range mt.DynamicElements {
			slice[v.Index+1] = "[" + v.Name + "]"
		}
		ns = strings.Join(slice, fc)
	}
	return ns
}
