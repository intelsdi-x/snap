package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/intelsdilabs/pulse/control"
)

var (
	// PulsePath is the local file path to a pulse build
	PulsePath = os.Getenv("PULSE_PATH")
	// CollectorPath is the path to collector plugins within a pulse build
	CollectorPath = path.Join(PulsePath, "plugin", "collector")
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

	pluginControl := control.New()
	pluginControl.Start()
	defer pluginControl.Stop()

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

	for {
		time.Sleep(time.Second * 1)
	}
	// err := pluginControl.Load(collectorPath)

}
