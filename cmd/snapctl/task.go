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

func stringValToInt(val string) (int, error) {
	// parse the input (string) value as an integer (log a Fatal
	// error if we cannot parse the value as an integer)
	parsedField, err := strconv.Atoi(val)
	if err != nil {
		splitErr := strings.Split(err.Error(), ": ")
		errStr := splitErr[len(splitErr)-1]
		// return a value of zero and the error encountered during string parsing
		return 0, fmt.Errorf("Value '%v' cannot be parsed as an integer (%v)", val, errStr)
	}
	// return the integer equivalent of the input string and a nil error (indicating success)
	return int(parsedField), nil
}

// Parses the command-line parameters and uses them to override the underlying schedule
// in this task (if any) or set a schedule for that task (if one is not already defined)
func (t *task) setWindowedSchedule(start *time.Time, stop *time.Time, duration *time.Duration) error {
	// if this task does not have an existing schedule, create an empty
	// windowed schedule
	if t.Schedule == nil {
		t.Schedule = &client.Schedule{
			Type: "windowed",
		}
	} else if t.Schedule.Type == "" {
		// else if there is an empty schedule already defined for this task,
		// set the type for that schedule to 'windowed'
		t.Schedule.Type = "windowed"
	} else if t.Schedule.Type != "windowed" {
		// else if the task's existing schedule is not a 'windowed' schedule,
		// then return an error
		return fmt.Errorf("Usage error (schedule type mismatch); cannot replace existing schedule of type '%v' with a new, 'windowed' schedule", t.Schedule.Type)
	}
	// grab the task schedule's start/stop time (will use these later to fill in a
	// missing boundary value, if we find one missing in our input arguments)
	scheduleStart := t.Schedule.StartTime
	scheduleStop := t.Schedule.StopTime
	// if a duration was passed in, determine the start and stop times for our new
	// 'windowed' schedule from the input parameters
	if duration != nil {
		// if start and stop were both defined, then return an error (since all three parameters cannot be used
		// to define the 'windowed' schedule)
		if start != nil && stop != nil {
			return fmt.Errorf("Usage error (too many parameters); only two of the parameters that define the window (start time, stop time and duration) can be specified for a 'windowed' schedule")
		}
		// if start is set and stop is not then use duration to create stop
		if start != nil && stop == nil {
			newStop := start.Add(*duration)
			t.Schedule.StartTime = start
			t.Schedule.StopTime = &newStop
		} else if stop != nil && start == nil {
			// else if stop is set and start is not then use duration to create start
			newStart := stop.Add(*duration * -1)
			t.Schedule.StartTime = &newStart
			t.Schedule.StopTime = stop
		} else {
			// else, the start and stop are both undefined but a duration was passed in,
			// so use the current date/time (plus the 'createTaskNowPad' value) as the
			// start date/time and construct a stop date/time from that start date/time
			// and the duration
			newStart := time.Now().Add(createTaskNowPad)
			newStop := newStart.Add(*duration)
			t.Schedule.StartTime = &newStart
			t.Schedule.StopTime = &newStop
		}
		// return a nil error (indicating success)
		return nil
	} else if (start == nil && scheduleStart == nil) || (stop == nil && scheduleStop == nil) {
		// else if the duration is undefined and either the start or stop date/time is missing (and the
		// corresponding start/stop date/time the schedule for this task is also undefined) then throw
		// an error (since we can't construct a complete window)
		return fmt.Errorf("Usage error (missing boundary); both boundaries (the start and stop date-time) must be specified for a 'windowed' schedule when no duration is specified")
	}
	// if new start and stop date/times were both specified, then use them to define a new 'windowed'
	// schedule (provided the start is less than the stop)
	if start != nil && stop != nil {
		if start.After(*stop) {
			return fmt.Errorf("Usage error (start after stop); the start date-time (%v) must be before the stop date-time (%v)", start, stop)
		}
		t.Schedule.StartTime = start
		t.Schedule.StopTime = stop
	} else if start != nil {
		// otherwise, if only the start date/time was specified, use it to replace the current schedule's
		// start date/time (provided it is before the current schedule's stop date/time)
		if !start.Before(*(t.Schedule.StopTime)) {
			return fmt.Errorf("Usage error (start after existing stop); the new start date-time (%v) must be before the existing stop date-time (%v)", start, t.Schedule.StopTime)
		}
		t.Schedule.StartTime = start
	} else if stop != nil {
		// otherwise, if only the stop date/time was specified, use it to replace the current schedule's
		// stop date/time (provided it is after the current schedule's start date/time)
		if !stop.After(*(t.Schedule.StartTime)) {
			return fmt.Errorf("Usage error (stop before existing start); the new stop date-time (%v) must be after the existing start date-time (%v)", stop, t.Schedule.StartTime)
		}
		t.Schedule.StopTime = stop
	}
	// return a nil error (indicating success)
	return nil
}

// parse the command-line options and use them to setup a new schedule for this task
func (t *task) setScheduleFromCliOptions(ctx *cli.Context) error {
	// if there is no schedule associated with this task, create an empty one
	if t.Schedule == nil {
		t.Schedule = &client.Schedule{}
	}
	// check the start, stop, and duration values to see if we're looking at a windowed schedule (or not)
	// first, get the parameters that define the windowed schedule
	start := mergeDateTime(
		strings.ToUpper(ctx.String("start-time")),
		strings.ToUpper(ctx.String("start-date")),
	)
	stop := mergeDateTime(
		strings.ToUpper(ctx.String("stop-time")),
		strings.ToUpper(ctx.String("stop-date")),
	)
	// Grab the duration string (if one was passed in) and parse it
	durationStr := ctx.String("duration")
	var duration *time.Duration
	if ctx.IsSet("duration") || durationStr != "" {
		d, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("Bad duration format; %v", err)
		}
		duration = &d
	}
	// Grab the interval for the schedule (if one was provided); we will use this
	// string to determine the type of schedule ('simple' or 'cron') and to update
	// the interval of the existing schedule (note that an interval value was
	// not passed in and there is no interval defined for the schedule associated
	// with this task, it's an error)
	interval := ctx.String("interval")
	if !ctx.IsSet("interval") && interval == "" && t.Schedule.Interval == "" {
		return fmt.Errorf("Usage error (missing interval value); when constructing a new task schedule, an interval must be provided")
	}
	// if a start, stop, or duration value was provided, then the CLI options are being used to
	// specify a new 'windowed' schedule for this task, so use them to replace the 'windowed' schedule
	// in the current task (if one exists); note that it is an error to try to replace an existing
	// sechedule with a new schedule of a different type, so throw an error if that is the case
	if start != nil || stop != nil || duration != nil {
		t.setWindowedSchedule(start, stop, duration)
	} else if interval != "" {
		// otherwise, we're looking at either a 'simple' schedule or a 'cron' schedule;
		// the type we're looking at is determined by the interval value passed in via the CLI
		isCron := false
		_, err := time.ParseDuration(interval)
		if err != nil {
			// try interpreting interval as cron entry
			_, e := cron.Parse(interval)
			if e != nil {
				return fmt.Errorf("Bad interval format: cannot parse interval value '%v' either as a duration or a cron entry\n", interval)
			}
			isCron = true
		}
		// We're looking at either a 'simple' or a 'cron' schedule, so determine which
		// from the 'isCron' flag set, above, and set the schedule type appropriately
		// (note it is an error to attempt to replace an existing schedule with a schedule
		// of a different type, so thron an error if that is the case)
		schedType := "simple"
		if isCron {
			// make sure the current schedule type (if there is one) matches
			if t.Schedule.Type != "" && t.Schedule.Type != "cron" {
				return fmt.Errorf("Usage error; cannot replace existing schedule of type '%v' with a new, 'cron' schedule", t.Schedule.Type)
			}
			schedType = "cron"
		} else if t.Schedule.Type != "" && t.Schedule.Type != "simple" {
			// make sure the current schedule type (if there is one) matches
			return fmt.Errorf("Usage error; cannot replace existing schedule of type '%v' with a new, 'simple' schedule", t.Schedule.Type)
		}
		// set the type (if it's not already set) and interval for the schedule associated with this task
		if t.Schedule.Type == "" {
			t.Schedule.Type = schedType
		}
		t.Schedule.Interval = interval
	}
	// return a nil error (indicating success)
	return nil
}

// merge the command-line options into the current task
func (t *task) mergeCliOptions(ctx *cli.Context) error {
	// set the name of the task (if a 'name' was provided in the CLI options)
	name := ctx.String("name")
	if ctx.IsSet("name") || name != "" {
		t.Name = name
	}
	// set the deadline of the task (if a 'deadline' was provided in the CLI options)
	deadline := ctx.String("deadline")
	if ctx.IsSet("deadline") || deadline != "" {
		t.Deadline = deadline
	}
	// set the MaxFailures for the task (if a 'max-failures' value was provided in the CLI options)
	maxFailuresStrVal := ctx.String("max-failures")
	if ctx.IsSet("max-failures") || maxFailuresStrVal != "" {
		maxFailures, err := stringValToInt(maxFailuresStrVal)
		if err != nil {
			return err
		}
		t.MaxFailures = maxFailures
	}
	// shouldn't ever happen, but...
	if t.Version != 1 {
		return fmt.Errorf("Invalid version provided while creating task")
	}
	// set the schedule for the task from the CLI options (and return the results
	// of that method call, indicating whether or not an error was encountered while
	// setting up that schedule)
	return t.setScheduleFromCliOptions(ctx)
}

func createTaskUsingTaskManifest(ctx *cli.Context) error {
	// get the task manifest file to use
	path := ctx.String("task-manifest")
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)
	if e != nil {
		return fmt.Errorf("File error [%s]- %v\n", ext, e)
	}

	// create an empty task struct and unmarshal the contents of the file into that object
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

	// merge any CLI optiones specified by the user (if any) into the current task;
	// if an error is encountered, return it
	if err := t.mergeCliOptions(ctx); err != nil {
		return err
	}

	// and use the resulting struct to create a new task
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
	// Get the workflow manifest filename from the command-line
	path := ctx.String("workflow-manifest")
	ext := filepath.Ext(path)
	file, e := ioutil.ReadFile(path)

	// check to make sure that an interval was specified using the appropriate command-line flag
	interval := ctx.String("interval")
	if !ctx.IsSet("interval") && interval != "" {
		return fmt.Errorf("Workflow manifest requires that an interval be set via a command-line flag.")
	}
	if e != nil {
		return fmt.Errorf("File error [%s]- %v\n", ext, e)
	}

	// and unmarshal the contents of the workflow manifest file into a local workflow map
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

	// create a dummy task
	t := task{}

	// fill in the details for that task from the command-line arguments passed in by the user;
	// if an error is encountered, return it
	if err := t.mergeCliOptions(ctx); err != nil {
		return err
	}

	// and use the resulting struct (along with the workflow map we constructed, above) to create a new task
	r := pClient.CreateTask(t.Schedule, wf, t.Name, t.Deadline, !ctx.IsSet("no-start"), t.MaxFailures)
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
