package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/intelsdilabs/pulse/control"
	"github.com/intelsdilabs/pulse/schedule"
)

// Pulse Flags for command line
var version = flag.Bool("version", false, "print Pulse version")
var maxProcs = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Default is use all cores.")

var gitversion string

type CoreModule interface {
	Start() error
	Stop()
}

//type PulseControl struct {
//	pluginController CoreModule
//	scheduler        CoreModule
//}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("Pulse version:", gitversion)
		os.Exit(0)
	}

	setMaxProcs()

	//pulseControl := &PulseControl{}

	c := control.New()
	s := schedule.New()
	s.MetricManager = c

	//  Start our modules
	if err := startModule("Plugin Controller", c); err != nil {
		printErrorAndExit("Plugin Controller", err)
	}
	if err := startModule("Scheduler", s); err != nil {
		if c.Started {
			c.Stop()
		}
		printErrorAndExit("Scheduler", err)
	}

	startInterruptHandling(c, s)
	select {} //run forever and ever
}

func setMaxProcs() {
	var _maxProcs int
	numProcs := runtime.NumCPU()
	if *maxProcs <= 0 {
		if *maxProcs < 0 {
			log.Printf("WARNING: max_procs set to less than zero. Setting GOMAXPROCS to %v", numProcs)
		}
		_maxProcs = numProcs
	} else if *maxProcs > numProcs {
		log.Printf("WARNING: Not allowed to set GOMAXPROCS above number of processors in system. Setting GOMAXPROCS to %v", numProcs)
		_maxProcs = numProcs
	} else {
		_maxProcs = *maxProcs
	}
	log.Printf("Setting GOMAXPROCS to %v\n", _maxProcs)
	runtime.GOMAXPROCS(_maxProcs)

	//Verify setting worked
	actualNumProcs := runtime.GOMAXPROCS(0)
	if actualNumProcs != _maxProcs {
		log.Printf("WARNING: Specified max procs of %v but using %v", _maxProcs, actualNumProcs)
	}
}

func startModule(name string, m CoreModule) error {
	log.Printf("Starting Pulse Agent %s module", name)
	return m.Start()
}

func printErrorAndExit(name string, err error) {
	log.Println("ERROR:", err)
	log.Println("ERROR: Error starting Pulse Agent %s module. Exiting now.")
	os.Exit(1)
}

func startInterruptHandling(modules ...CoreModule) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	//Let's block until someone tells us to quit
	go func() {
		sig := <-c
		log.Println("Stopping Pulse Agent modules")
		for _, m := range modules {
			m.Stop()
		}
		log.Printf("Exiting given signal: %v", sig)
		os.Exit(0)
	}()
}
