/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015,2016 Intel Corporation

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/robfig/cron"

	"github.com/ghodss/yaml"
)

var (
	// padding to picking a time to start a "NOW" task
	createTaskNowPad = time.Second * 1
	timeParseFormat  = "3:04PM"
	dateParseFormat  = "1-02-2006"
	unionParseFormat = timeParseFormat + " " + dateParseFormat
)

// Constants used to truncate task hit and miss counts
// e.g. 1K(10^3), 1M(10^6, 1G(10^9) etc (not 1024^#). We do not
// use units larger than Gb to support 32 bit compiles.
const (
	K = 1000
	M = 1000 * K
	G = 1000 * M
)

func trunc(n int) string {
	var u string

	switch {
	case n >= G:
		u = "G"
		n /= G
	case n >= M:
		u = "M"
		n /= M
	case n >= K:
		u = "K"
		n /= K
	default:
		return strconv.Itoa(n)
	}
	return strconv.Itoa(n) + u
}

type task struct {
	Version     int
	Schedule    *client.Schedule
	Workflow    *wmap.WorkflowMap
	Name        string
	Deadline    string
	MaxFailures int `json:"max-failures"`
}

func createTask(ctx *cli.Context) error {
	var err error
	if ctx.IsSet("task-manifest") {
		fmt.Println("Using task manifest to create task")
		err = createTaskUsingTaskManifest(ctx)
	} else if ctx.IsSet("workflow-manifest") {
		fmt.Println("Using workflow manifest to create task")
		err = createTaskUsingWFManifest(ctx)
	} else {
		return newUsageError("Must provide either --task-manifest or --workflow-manifest arguments", ctx)
	}
	return err
}

func createTaskUsingTaskManifest(ctx *cli.Context) error {
	path := ctx.String("task-manifest")
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return fmt.Errorf("File error [%s]- %v\n", ext, e)
	}

	t := task{}
	switch ext {
	case ".yaml", ".yml":
		e = yaml.Unmarshal(file, &t)
		if e != nil {
			return fmt.Errorf("Error parsing YAML file input - %v\n", e)
		}
	case ".json":
		e = json.Unmarshal(file, &t)
		if e != nil {
			return fmt.Errorf("Error parsing JSON file input - %v\n", e)
		}
	default:
		return fmt.Errorf("Unsupported file type %s\n", ext)
	}

	t.Name = ctx.String("name")
	if t.Version != 1 {
		return fmt.Errorf("Invalid version provided")
	}

	// If the number of failures does not specific, default value is 10
	if t.MaxFailures == 0 {
		fmt.Println("If the number of maximum failures is not specified, use default value of", DefaultMaxFailures)
		t.MaxFailures = DefaultMaxFailures
	}

	r := pClient.CreateTask(t.Schedule, t.Workflow, t.Name, t.Deadline, !ctx.IsSet("no-start"), t.MaxFailures)

	if r.Err != nil {
		errors := strings.Split(r.Err.Error(), " -- ")
		errString := "Error creating task:"
		for _, err := range errors {
			errString += fmt.Sprintf("%v\n", err)
		}
		return fmt.Errorf(errString)
	}
	fmt.Println("Task created")
	fmt.Printf("ID: %s\n", r.ID)
	fmt.Printf("Name: %s\n", r.Name)
	fmt.Printf("State: %s\n", r.State)

	return nil
}

func createTaskUsingWFManifest(ctx *cli.Context) error {
	// Get the workflow
	path := ctx.String("workflow-manifest")
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)

	if !ctx.IsSet("interval") && !ctx.IsSet("i") {
		return fmt.Errorf("Workflow manifest requires interval to be set via flag.")
	}
	if e != nil {
		return fmt.Errorf("File error [%s]- %v\n", ext, e)
	}

	var wf *wmap.WorkflowMap
	switch ext {
	case ".yaml", ".yml":
		// e = yaml.Unmarshal(file, &t)
		wf, e = wmap.FromYaml(file)
		if e != nil {
			return fmt.Errorf("Error parsing YAML file input - %v\n", e)
		}
	case ".json":
		wf, e = wmap.FromJson(file)
		// e = json.Unmarshal(file, &t)
		if e != nil {
			return fmt.Errorf("Error parsing JSON file input - %v\n", e)
		}
	}
	// Get the task name
	name := ctx.String("name")
	// Get the interval
	isCron := false
	i := ctx.String("interval")
	_, err := time.ParseDuration(i)
	if err != nil {
		// try interpreting interval as cron entry
		_, e := cron.Parse(i)
		if e != nil {
			return fmt.Errorf("Bad interval format:\nfor simple schedule: %v\nfor cron schedule: %v\n", err, e)
		}
		isCron = true
	}

	// Deadline for a task
	dl := ctx.String("deadline")
	maxFailures := ctx.Int("max-failures")

	var sch *client.Schedule
	// None of these mean it is a simple schedule
	if !ctx.IsSet("start-date") && !ctx.IsSet("start-time") && !ctx.IsSet("stop-date") && !ctx.IsSet("stop-time") {
		// Check if duration was set
		if ctx.IsSet("duration") && !isCron {
			d, err := time.ParseDuration(ctx.String("duration"))
			if err != nil {
				return fmt.Errorf("Bad duration format:\n%v\n", err)
			}
			start := time.Now().Add(createTaskNowPad)
			stop := start.Add(d)
			sch = &client.Schedule{
				Type:      "windowed",
				Interval:  i,
				StartTime: &start,
				StopTime:  &stop,
			}
		} else {
			// No start or stop and no duration == simple schedule
			t := "simple"
			if isCron {
				// It's a cron schedule, ignore "duration" if set
				t = "cron"
			}
			sch = &client.Schedule{
				Type:     t,
				Interval: i,
			}
		}
	} else {
		// We have some form of windowed schedule
		start := mergeDateTime(
			strings.ToUpper(ctx.String("start-time")),
			strings.ToUpper(ctx.String("start-date")),
		)
		stop := mergeDateTime(
			strings.ToUpper(ctx.String("stop-time")),
			strings.ToUpper(ctx.String("stop-date")),
		)

		// Use duration to create missing start or stop
		if ctx.IsSet("duration") {
			d, err := time.ParseDuration(ctx.String("duration"))
			if err != nil {
				return fmt.Errorf("Bad duration format:\n%v\n", err)
			}
			// if start is set and stop is not then use duration to create stop
			if start != nil && stop == nil {
				t := start.Add(d)
				stop = &t
			}
			// if stop is set and start is not then use duration to create start
			if stop != nil && start == nil {
				t := stop.Add(d * -1)
				start = &t
			}
		}
		sch = &client.Schedule{
			Type:      "windowed",
			Interval:  i,
			StartTime: start,
			StopTime:  stop,
		}
	}
	// Create task
	r := pClient.CreateTask(sch, wf, name, dl, !ctx.IsSet("no-start"), maxFailures)
	if r.Err != nil {
		errors := strings.Split(r.Err.Error(), " -- ")
		errString := "Error creating task:"
		for _, err := range errors {
			errString += fmt.Sprintf("%v\n", err)
		}
		return fmt.Errorf(errString)
	}
	fmt.Println("Task created")
	fmt.Printf("ID: %s\n", r.ID)
	fmt.Printf("Name: %s\n", r.Name)
	fmt.Printf("State: %s\n", r.State)

	return nil
}

func mergeDateTime(tm, dt string) *time.Time {
	reTm := time.Now().Add(createTaskNowPad)
	if dt == "" && tm == "" {
		return nil
	}
	if dt != "" {
		t, err := time.Parse(dateParseFormat, dt)
		if err != nil {
			fmt.Printf("Error creating task:\n%v\n", err)
			os.Exit(1)
		}
		reTm = t
	}

	if tm != "" {
		_, err := time.ParseInLocation(timeParseFormat, tm, time.Local)
		if err != nil {
			fmt.Printf("Error creating task:\n%v\n", err)
			os.Exit(1)
		}
		reTm, err = time.ParseInLocation(unionParseFormat, fmt.Sprintf("%s %s", tm, reTm.Format(dateParseFormat)), time.Local)
		if err != nil {
			fmt.Printf("Error creating task:\n%v\n", err)
			os.Exit(1)
		}
	}
	return &reTm
}

func listTask(ctx *cli.Context) error {
	tasks := pClient.GetTasks()
	if tasks.Err != nil {
		return fmt.Errorf("Error getting tasks:\n%v\n", tasks.Err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	printFields(w, false, 0,
		"ID",
		"NAME",
		"STATE",
		"HIT",
		"MISS",
		"FAIL",
		"CREATED",
		"LAST FAILURE",
	)
	for _, task := range tasks.ScheduledTasks {
		printFields(w, false, 0,
			task.ID,
			task.Name,
			task.State,
			trunc(task.HitCount),
			trunc(task.MissCount),
			trunc(task.FailedCount),
			task.CreationTime().Format(unionParseFormat),
			task.LastFailureMessage,
		)
	}
	w.Flush()

	return nil
}

func watchTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	verbose := ctx.Bool("verbose")
	id := ctx.Args().First()
	r := pClient.WatchTask(id)
	if r.Err != nil {
		return fmt.Errorf("%v", r.Err)
	}
	fmt.Printf("Watching Task (%s):\n", id)

	// catch interrupt so we signal the server we are done before exiting
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	var lines int
	go func() {
		<-c
		fmt.Printf("%sStopping task watch\n", strings.Repeat("\n", lines))
		r.Close()
		return
	}()

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fields := []interface{}{"NAMESPACE", "DATA", "TIMESTAMP"}
	if verbose {
		fields = append(fields, "TAGS")
	}
	printFields(w, false, 0, fields...)

	// Loop listening to events
	for {
		select {
		case e := <-r.EventChan:
			switch e.EventType {
			case "metric-event":
				sort.Sort(e.Event)
				var extra int
				for _, event := range e.Event {
					fmt.Printf("\033[0J")
					eventFields := []interface{}{
						event.Namespace,
						event.Data,
						event.Timestamp,
					}
					if !verbose {
						printFields(w, false, 0, eventFields...)
						continue
					}
					tags := sortTags(event.Tags)
					if len(tags) <= 3 {
						eventFields = append(eventFields, strings.Join(tags, ", "))
						printFields(w, false, 0, eventFields...)
						continue
					}
					for i := 0; i < len(tags); i += 3 {
						tagSlice := tags[i:min(i+3, len(tags))]
						if i == 0 {
							eventFields = append(eventFields, strings.Join(tagSlice, ", ")+",")
							printFields(w, false, 0, eventFields...)
							continue
						}
						extra += 1
						if i+3 > len(tags) {
							printFields(w, false, 0,
								"",
								"",
								"",
								strings.Join(tagSlice, ", "),
							)
							continue
						}
						printFields(w, false, 0,
							"",
							"",
							"",
							strings.Join(tagSlice, ", ")+",",
						)

					}
				}
				lines = len(e.Event) + extra
				fmt.Fprintf(w, "\033[%dA\n", lines+1)
				w.Flush()
			default:
				fmt.Printf("%s[%s]\n", strings.Repeat("\n", lines), e.EventType)
			}

		case <-r.DoneChan:
			return nil
		}
	}

}

func startTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()
	r := pClient.StartTask(id)
	if r.Err != nil {
		if strings.Contains(r.Err.Error(), "Task is already running.") {
			fmt.Println("Task is already running")
			fmt.Printf("ID: %s\n", id)
			return nil
		}
		return fmt.Errorf("Error starting task:\n%v\n", r.Err)
	}
	fmt.Println("Task started:")
	fmt.Printf("ID: %s\n", r.ID)

	return nil
}

func stopTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()
	r := pClient.StopTask(id)
	if r.Err != nil {
		if strings.Contains(r.Err.Error(), "Task is already stopped.") {
			fmt.Println("Task is already stopped")
			fmt.Printf("ID: %s\n", id)
			os.Exit(0)
		}
		return fmt.Errorf("Error stopping task:\n%v\n", r.Err)
	}
	fmt.Println("Task stopped:")
	fmt.Printf("ID: %s\n", r.ID)

	return nil
}

func removeTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()
	r := pClient.RemoveTask(id)
	if r.Err != nil {
		return fmt.Errorf("Error stopping task:\n%v\n", r.Err)
	}
	fmt.Println("Task removed:")
	fmt.Printf("ID: %s\n", r.ID)

	return nil
}

func exportTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}
	id := ctx.Args().First()
	task := pClient.GetTask(id)
	if task.Err != nil {
		return fmt.Errorf("Error exporting task:\n%v\n", task.Err)
	}
	tb, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("Error exporting task:\n%v\n", err)
	}
	fmt.Println(string(tb))
	return nil
}

func enableTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return newUsageError("Incorrect usage", ctx)
	}

	id := ctx.Args().First()
	r := pClient.EnableTask(id)
	if r.Err != nil {
		return fmt.Errorf("Error enabling task:\n%v\n", r.Err)
	}
	fmt.Println("Task enabled:")
	fmt.Printf("ID: %s\n", r.ID)
	return nil
}

func sortTags(tags map[string]string) []string {
	var tagSlice []string
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tagSlice = append(tagSlice, k+"="+tags[k])
	}
	return tagSlice
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
