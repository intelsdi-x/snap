// This modules abstracts communication with external facter program (command)
package facter

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"time"
)

type cmdConfig struct {
	// where to find facter executable
	// tip: somehow golang is able to find executable within PATH, so just a name is enough
	// even if facter is just a ruby script
	executable string
	// default options passed to facter to get json output
	// overriden during tests to get non parseable output
	options []string
}

// overriden durning tests to simulate error
func newDefaultCmdConfig() *cmdConfig {
	return &cmdConfig{
		// tip: somehow golang is able to find executable within PATH, so just a name is enough
		executable: "facter",
		options:    []string{"--json"},
	}
}

// helper type to deal with json parsing
type stringmap map[string]interface{}

// get facts from facter (external command)
// returns all keys if none requested
// if cmdOptions is nil, use defaults "facter --json"
func getFacts(keys []string, facterTimeout time.Duration, cmdConfig *cmdConfig) (*stringmap, *time.Time, error) {

	// nil means use defaults
	if cmdConfig == nil {
		cmdConfig = newDefaultCmdConfig()
	}

	// default options -
	args := make([]string, len(cmdConfig.options))
	copy(args, cmdConfig.options)
	// is this more elegant that than this:
	// > args := append([]string{}, cmdConfig.options...)
	args = append(args, keys...)

	// execute command and capture the output
	jobCompletedChan := make(chan struct{})
	timeoutChan := time.After(facterTimeout)

	var err error
	output := make([]byte, 0, 1024)

	go func(jobCompletedChan chan<- struct{}, output *[]byte, err *error) {
		*output, *err = exec.Command(cmdConfig.executable, args...).Output()
		jobCompletedChan <- struct{}{}
	}(jobCompletedChan, &output, &err)

	select {
	case <-timeoutChan:
		return nil, nil, errors.New("Facter plugin: fact gathering timeout")
	case <-jobCompletedChan:
		// success
	}

	if err != nil {
		log.Println("exec returned ", err.Error())
		return nil, nil, err
	}

	timestamp := time.Now()

	var facterMap stringmap
	err = json.Unmarshal(output, &facterMap)
	if err != nil {
		log.Println("Unmarshal failed ", err.Error())
		return nil, nil, err
	}
	return &facterMap, &timestamp, nil
}
