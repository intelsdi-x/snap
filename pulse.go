package main

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/intelsdilabs/pulse/control"
)

var (
	PulsePath     = os.Getenv("PULSE_PATH")
	CollectorPath = path.Join(PulsePath, "plugin", "collector") // todo figure out whether env or determined

	// TODO
	PULSE_COLLECTOR_PATH = os.Getenv("PULSE_COLLECTOR_PATH")
	PULSE_PUBLISHER_PATH = os.Getenv("PULSE_PUBLISHER_PATH")
)

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func isExecutable(p os.FileMode) bool {
	return (p & 0111) != 0
}

func main() {
	if PulsePath == "" {
		log.Fatalln("PULSE_PATH not set")
		os.Exit(1)
	}

	log.Println("Startup")
	// TODO ERROR missing PULSE_PATH
	// fmt.Println(PulsePath)

	pluginControl := control.Control()

	// fmt.Println(pluginControl)
	// fmt.Println(CollectorPath)

	m, err := filepath.Glob(CollectorPath + "/pulse-collector-*")
	checkError(err)
	for _, d := range m {
		f, err := os.Stat(d)
		checkError(err)
		// Ignore directories
		if f.Mode().IsDir() {
			continue
		}
		// Ignore file without executable permission
		if !isExecutable(f.Mode().Perm()) {
			log.Printf("The plugin [%s] is not executable\n", d)
			continue
		}
		pluginControl.Load(d)
	}
	// err := pluginControl.Load(collectorPath)

}
