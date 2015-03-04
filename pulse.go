package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	//	"path"
	//	"path/filepath"
	//	"time"

	//	"github.com/intelsdilabs/pulse/control"
)

var version = flag.Bool("version", false, "print Pulse version")
var maxProcs = flag.Int("max_procs", 0, "max number of CPUs that can be used simultaneously. Default is use all cores.")

type PulseController struct{} //Will hold what modules start and stop

func main() {
	//TODO: Parse startup arguments for pulse
	flag.Parse()
	if *version {
		//TODO: Pass in version during build
		fmt.Println("Pulse version 1.0.0-alpha\n")
		os.Exit(0)
	}

	pc := &PulseController{}
	setMaxProcs()

	//TODO: Start pulse modules

	startInterruptHandling(pc)
	select {} //run forever and ever
}

func setMaxProcs() {
	var _maxProcs int
	numProcs := runtime.NumCPU()
	if *maxProcs == 0 {
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
		log.Println("WARNING: Specified max procs of %v but using %v", _maxProcs, actualNumProcs)
	}
}

func startInterruptHandling(pc *PulseController) {
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

//var (
// PulsePath is the local file path to a pulse build
//	PulsePath = os.Getenv("PULSE_PATH")
// CollectorPath is the path to collector plugins within a pulse build
//	CollectorPath = path.Join(PulsePath, "plugin", "collector")
//)

//func checkError(e error) {
//	if e != nil {
//		panic(e)
//	}
//}

//func isExecutable(p os.FileMode) bool {
//	return (p & 0111) != 0
//}

//func main() {
//	if PulsePath == "" {
//		log.Fatalln("PULSE_PATH not set")
//		os.Exit(1)
//	}
//
//	log.Println("Startup")
// TODO ERROR missing PULSE_PATH
// fmt.Println(PulsePath)

//	pluginControl := control.New()
//	pluginControl.Start()
//	defer pluginControl.Stop()

// fmt.Println(pluginControl)
// fmt.Println(CollectorPath)

//	m, err := filepath.Glob(CollectorPath + "/pulse-collector-*")
//	checkError(err)
//	for _, d := range m {
//		f, err := os.Stat(d)
//		checkError(err)
//		// Ignore directories
//		if f.Mode().IsDir() {
//			continue
//		}
//		// Ignore file without executable permission
//		if !isExecutable(f.Mode().Perm()) {
//			log.Printf("The plugin [%s] is not executable\n", d)
//			continue
//		}
//		pluginControl.Load(d)
//	}
//
//	for {
//		time.Sleep(time.Second * 1)
//	}
//	// err := pluginControl.Load(collectorPath)
//
//}
