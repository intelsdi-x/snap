package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/intelsdilabs/pulse/control"
)

var (
	// PulsePath is the local file path to a pulse build
	PulsePath = os.Getenv("PULSE_PATH")
	// CollectorPath is the path to collector plugins within a pulse build
	CollectorPath = path.Join(PulsePath, "plugin", "collector")
)

// Pulse Flags for command line
var version = flag.Bool("version", false, "print Pulse version")
var maxProcs = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Default is use all cores.")
var pluginController = flag.Bool("plugin_controller", false, "Enable Pulse Plugin Controller")

var gitversion string

type CoreModule interface {
	Start()
	Stop()
}

type PulseControl struct {
	pluginController CoreModule
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("Pulse version:", gitversion)
		os.Exit(0)
	}

	setMaxProcs()

	pulseControl := &PulseControl{}

	if *pluginController {
		pulseControl.pluginController = control.New()
	}
	//TODO: Start pulse modules
	//TODO: Set module order for starting and stopping
	startInterruptHandling()
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

func startInterruptHandling() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	//Let's block until someone tells us to quit
	go func() {
		sig := <-c
		//TODO: Stop pulse modules on exit signals
		log.Printf("Exiting given signal: %v", sig)
		os.Exit(0)
	}()
}
