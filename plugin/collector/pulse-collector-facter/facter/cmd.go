// This modules abstracts communication with external facter program (command)
package facter

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"time"
)

// configuration variables
var (
	// where to find facter executable
	// tip: somehow golang is able to find executable within PATH, so just a name is enough
	// even if facter is just a ruby script
	// overriden durning tests
	facter_executable = "facter"
)

// helper type to deal with json that stores last update moment
// for a given fact
type fact struct {
	value      interface{}
	lastUpdate time.Time
}

// helper type to deal with json parsing
type stringmap map[string]interface{}

// get facts from facter (external command)
// returns all keys if none requested
func getFacts(keys []string, facterTimeout time.Duration) (*stringmap, *time.Time, error) {

	var timestamp time.Time

	// default options
	args := []string{"--json"}
	args = append(args, keys...)

	// execute command and capture the output
	jobCompletedChan := make(chan struct{})
	timeoutChan := time.After(facterTimeout)

	var err error
	output := make([]byte, 0, 1024)

	go func(jobCompletedChan chan<- struct{}, output *[]byte, err *error) {
		*output, *err = exec.Command(facter_executable, args...).Output()
		jobCompletedChan <- struct{}{}
	}(jobCompletedChan, &output, &err)

	select {
	case <-timeoutChan:
		return nil, nil, errors.New("Facter plugin: fact gathering timeout")
	case <-jobCompletedChan:
		// success
		break
	}

	if err != nil {
		log.Println("exec returned " + err.Error())
		return nil, nil, err
	}
	timestamp = time.Now()

	var facterMap stringmap
	err = json.Unmarshal(output, &facterMap)
	if err != nil {
		log.Println("Unmarshal failed " + err.Error())
		return nil, nil, err
	}
	return &facterMap, &timestamp, nil
}
