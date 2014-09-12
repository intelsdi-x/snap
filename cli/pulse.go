// Intel Pulse® Telemetry Agent
// This is the agent CLI

/*

pulse
		Command line args override config

		--disable-caching
		--cache-ttl <TTL>
		--log-level
		--log-file

		list
				--format json(j), yaml(y), list(l)[default]

		get
				--host
						nweaver-intel.*
				--name
						**,steal
						*,disk
						cpu,**
				--truncate <LENGTH> [default 32 characters]
				--format json(j), yaml(y), list(l)[default]
				--collector collectd,facter

		server
				--live-brief
				--live-tasks
				--live-workers
				--live-all
				--web
				--access-user user
				--access-password password
				--config <FILEPATH>
				--workers <NUM OF WORKERS> (Defaults to core count)



*/

package main

import (
	"runtime"
	"github.com/lynxbat/pulse"
	"github.com/lynxbat/pulse/collection"
//	"github.com/lynxbat/pulse/server"
	"github.com/codegangsta/cli"
	"os"
	"fmt"
	"strings"
	"text/tabwriter"
	"encoding/json"
	"time"
	"reflect"
)

type MetricValueItem struct {
	Host, Collector string
	Last_update time.Time
	Namespace []string
	Value interface{}
}

type MetricListItem struct {
	Host string `json:host`
	Collector string `json:collector`
	LastUpdate time.Time `json:last_update`
	Namespace []string `json:namespace`
}

func main() {
	app := cli.NewApp()
	app.Name = "pulse"
	app.Usage = "Intel Pulse® Telemetry Agent"
	app.Commands = []cli.Command{
		{
			Name:      "list",
			Usage:     "list metrics for this agent",
			Flags:     []cli.Flag {
				cli.StringFlag{Name: "format, f", Value: "list", Usage: "json(j), yaml(y), list(l)[default]"},
				cli.BoolFlag{Name: "no-pretty-print", Usage: "Disables pretty printing of JSON"},
			},
			Action: func(c *cli.Context) {
				listCommand(c)
			},
		},
		{
			Name:      "get",
			Usage:     "Retrive metrics",
			Flags:		[]cli.Flag {
				cli.StringFlag{Name: "format, f", Value: "list", Usage: "json(j), yaml(y), list(l)[default]"},
				cli.BoolFlag{Name: "no-pretty-print", Usage: "Disables pretty printing of JSON"},
			},
			Action: func(c *cli.Context) {
				getCommand(c)
			},
		},
		{
			Name:      "server",
			Usage:     "Start Pulse in server mode",
			Flags:		[]cli.Flag {
				cli.BoolFlag{Name: "interactive, i", Usage: "Logs to STDOUT"},
				cli.IntFlag{Name: "maxcpu", Value: runtime.NumCPU(), Usage: fmt.Sprintf("Number of cores to run across. Default [%d]", runtime.NumCPU())},
				cli.BoolFlag{Name: "web, w", Usage: "Enables Web Server (REST API / status page)"},
				cli.StringFlag{Name: "access-user", Usage: "Enter a user to secure web server"},
				cli.StringFlag{Name: "access-password", Usage: "Enter a password to secure web server"},
				cli.StringFlag{Name: "config, c", Usage: "File path to pulse configuration file"},
			},
			Action: func(c *cli.Context) {
				serverCommand(c)
			},
		},
	}

	app.Run(os.Args)
}

func serverCommand(c *cli.Context) {
	agent.StartScheduler(c.Int("maxcpu"))





//	if c.Bool("web") {
//		s := new(server.Server)
//		s.Port = 3000
//		s.Start()
//	}
}

func getCommand(c *cli.Context) {
	switch c.String("format") {
	case "list", "l":
		PrintScreenMetricValues(agent.GetMetricValues())
	case "json", "j":
		if c.Bool("no-pretty-print") {
			PrintMetricValuesAsJSON(agent.GetMetricValues())
		} else {
			PrintMetricValuesAsJSON(agent.GetMetricValues())
		}
	case "yaml", "y":
		// Need to find good yaml lib that works without horrible licensing
		fmt.Println("NOT IMPLEMENTED")
	default:
		fmt.Printf("Invalid list format \"%v\"\n", c.String("format"))
		cli.ShowCommandHelp(c, "get")
	}
}

func listCommand(c *cli.Context) {
	switch c.String("format") {
	case "list", "l":
		PrintScreenMetricList(agent.GetMetricList())
	case "json", "j":
		if c.Bool("no-pretty-print") {
			PrintMetricsListAsJSON(agent.GetMetricList(), false)
		} else {
			PrintMetricsListAsJSON(agent.GetMetricList(), true)
		}
	case "yaml", "y":
		// Need to find good yaml lib that works without horrible licensing
		fmt.Println("NOT IMPLEMENTED")
	default:
		fmt.Printf("Invalid list format \"%v\"\n", c.String("format"))
		cli.ShowCommandHelp(c, "list")
	}
}

func PrintScreenMetricValues(metrics []collection.Metric) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)

	value_count := 0
	for _, metric := range metrics {
		for k := range metric.Values {
			val, type_ := parseValues(metric.Values[k])

			if len(val) > 32 {
				val = val[:32] + "..."
			}

			fmt.Fprintf(w, "\t%v\t%v\t%v\t%v\t%v\t%v\n", metric.Host, metric.GetNamespaceString(), k, val, type_, metric.Collector)
			value_count++
		}
	}
	w.Flush()

	fmt.Printf("\nTotal metric values: %d\n", value_count)
}

func PrintMetricValuesAsJSON(metrics []collection.Metric) {

	mlist := []MetricValueItem{}
	for _, metric := range metrics {
		for k := range metric.Values {
			mlist = append(mlist, MetricValueItem{metric.Host, metric.Collector, metric.LastUpdate, append(metric.Namespace, k), metric.Values[k]})
		}
	}
	b, e := json.MarshalIndent(mlist, "", " ")
	if e != nil {
		panic(e)
	}
	fmt.Println(string(b))
}

func PrintScreenMetricList(metrics []collection.Metric) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, ' ', 0)

	header := "Host" + "\t" + "Namespace" + "\t" + "Collector" + "\t" + "Last Update"
	fmt.Fprintln(w, header)
	for _, metric := range metrics {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n", metric.Host, strings.Join(metric.Namespace, "/"), metric.Collector, metric.LastUpdate.Format("2006/01/02 15:04:05"))
	}
	w.Flush()
}

func newMetricListItem(host, collector string, last_update time.Time, namespace []string) MetricListItem{
	return MetricListItem{host, collector, last_update, namespace}
}

func PrintMetricsListAsJSON(metrics []collection.Metric, pretty bool) {
	mlist := []MetricListItem{}
	for _, m :=  range metrics {
		mlist = append(mlist, MetricListItem{m.Host, m.Collector, m.LastUpdate, m.Namespace})
	}
	var b []byte
	if pretty {
		b, _ = json.MarshalIndent(mlist, "", " ")
	} else {
		b, _ = json.Marshal(mlist)
	}
	fmt.Println(string(b))
}

func parseValues(in_val interface{}) (string, string){
	var val, type_ string

	// Lots of conversion
	if v, ok := in_val.(string); ok {
		val = v
		type_ = "string"
	} else if v, ok := in_val.(float64); ok {
		val = fmt.Sprintf("%.6f", v)
		type_ = "float"
	} else if v, ok := in_val.(uint64); ok {
		val = fmt.Sprintf("%d", v)
		type_ = "integer"
	} else if v, ok := in_val.(int); ok {
		val = fmt.Sprintf("%d", v)
		type_ = "integer"
	} else if v, ok := in_val.(bool); ok {
		val = fmt.Sprintf("%t", v)
		type_ = "bool"
	} else if v, ok := in_val.([]uint64); ok {
		b, _ := json.Marshal(v)
		val = string(b)
		type_ = "json"
	} else if v, ok := in_val.(map[string]interface {}); ok {
		b, _ := json.Marshal(v)
		val = string(b)
		type_ = "json"
	} else if v, ok := in_val.([]string); ok {
		b, _ := json.Marshal(v)
		val = string(b)
		type_ = "json"
	} else {
		// Missed a type
		fmt.Printf("%v\n", reflect.TypeOf(in_val))
		panic(1)
	}
	return val, type_
}
