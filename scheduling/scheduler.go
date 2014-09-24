package scheduling

import (
	"../collection"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	//	tm "github.com/buger/goterm" // TODO move to pulse cli
	"os"
)

type SchedulerState int

const (
	Initialized SchedulerState = iota
	Starting
	Running
	Stopping
	Stopped
)

// Manages metric workers
type scheduler struct {
	MetricTasks []*MetricTask
	State       SchedulerState

	workerPoolCount   int
	workerPool        WorkerPool
	workerChan        chan work
	workerQuitChan    chan bool
	workerAckQuitChan chan bool
}

type schedule struct {
	Start    *time.Time
	Stop     *time.Time
	Interval time.Duration
	// Interval will be fixed to static division inside start-stop range. Optionally maybe have it based on from task start
}

type work struct {
	Collector string
	Metrics   []collection.Metric
}

func NewScheduler(initWorkerCount int) scheduler {
	s := scheduler{}
	s.State = Initialized
	s.workerPoolCount = initWorkerCount
	return s
}

func NewSchedule(duration time.Duration, times ...time.Time) schedule {
	s := schedule{}
	s.Interval = duration
	if len(times) > 0 {
		// Assign first as Start
		s.Start = &times[0]
		if len(times) > 1 {
			// Assign second as Stop
			s.Stop = &times[1]
		}
	}
	return s
}

type WorkerPool []*MetricWorker

func (m WorkerPool) Len() int {
	return len(m)
}

func (m WorkerPool) Less(i, j int) bool {
	return m[i].UUID() < m[j].UUID()
}

func (m WorkerPool) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (s *scheduler) Start() error {
	// Scheduler needs to know the types of collectors across all tasks (and when a new task is added)
	// Scheduler is responsible for:
	// * initializing new collectors
	// * closing collector
	//	collectorMap := collection.NewCollectorMap()
	for _, cType := range s.getCollectorTypes() {
		// For each c call NewCollectorByType(cType) and store in collectorMap[c]
		fmt.Printf("Creating collector type: %s\n", cType)
		//		collection.NewCollectorByType(cType)
	}

	os.Exit(0)

	s.State = Starting
	s.workerChan = make(chan work)
	s.workerQuitChan = make(chan bool)
	s.workerAckQuitChan = make(chan bool)

	fmt.Println("Scheduler started")

	// Check for problems
	if !s.HasTasks() {
		return errors.New("No tasks defined")
	}

	// Control concurrency across cores
	// Workers are 1:1 with go routines that call collectors
	// Workers pool is adjustable. Limiting the pool can ensure less overhead.
	s.workerPool = WorkerPool{}
	// Start printing of stats. TODO // convert to Stats method so printing is an op above scheduler
	go s.PrintStats()

	// Create and start workers
	for x := 0; x < s.workerPoolCount; x++ {
		mw := new(MetricWorker)
		go mw.Start(s.workerChan, s.workerQuitChan, s.workerAckQuitChan)
		s.workerPool = append(s.workerPool, mw)
	}

	// Create start spinning tasks
	for _, t := range s.MetricTasks {
		go t.Spin(s.workerChan)
	}
	return nil
}

// TODO remove
func (s *scheduler) PrintStats() {
	sort.Sort(s.workerPool)
	var lineCount int = 0
	for {
		if s.State == Stopped || s.State == Stopping {
			return
		}
		x := ""
		x += "\n"
		for i, t := range s.MetricTasks {
			if t.State() == "waiting" && t.HasStart() {
				x += fmt.Sprintf("  task [%d][%s] State: %s    Starts in: %0.2fsec    \n", i, t.Label, t.State(), t.TimeTillStart().Seconds())
			} else if t.State() == "stopped" {
				x += fmt.Sprintf("  task [%d][%s] State: %s                           \n", i, t.Label, t.State())
			} else {
				x += fmt.Sprintf("  task [%d][%s] State: %s    Drift: %0.3fsec        \n", i, t.Label, t.State(), t.Drift.Seconds())
			}
		}
		x += "\n"
		t_jobs, t_metrics := 0, 0
		for i, w := range s.workerPool {
			x += fmt.Sprintf("  worker [%d][%s]    Jobs: %d    Metrics: %d\n", i, w.UUID(), w.WorkReceived, w.MetricsWorked)
			t_jobs += w.WorkReceived
			t_metrics += w.MetricsWorked
		}
		x += "\n"
		x += fmt.Sprintf("Total Jobs: %d\nTotal Metrics: %d\n", t_jobs, t_metrics)
		x += "\n"
		if lineCount > 0 {
			fmt.Printf("\033[%dA", lineCount-1)
		}
		fmt.Print(x)
		lineCount = len(strings.Split(x, "\n"))
		time.Sleep(time.Millisecond * 500)
	}
}

func (s *scheduler) Stop() error {
	s.State = Stopping
	fmt.Print("\n\nStopping")
	for x := 0; x < len(s.workerPool); x++ {
		s.workerQuitChan <- true
	}
	x := 0
	for x < len(s.workerPool) {
		<-s.workerAckQuitChan
		x++
		// A little padding to let goroutine exit. Not needed but useful for debug.
		// In reality the worker goroutines are safe to be killed by os.Exit by now.
		time.Sleep(time.Millisecond * 100)
		fmt.Print(("."))
	}
	s.State = Stopped
	fmt.Println("\nStopped")
	return nil
}

func (s *scheduler) HasTasks() bool {
	return len(s.MetricTasks) > 0
}

func (s *scheduler) getCollectorTypes() []string {
	h := map[string]bool{}
	c := []string{}

	for _, t := range s.MetricTasks {
		for _, m := range t.Metrics {
			h[m.Collector] = true
		}
	}
	for k, _ := range h {
		c = append(c, k)
	}
	return c
}
