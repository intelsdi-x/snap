// This modules abstracts communication with external facter program (external binary)
package facter

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"time"
)

// helper type to deal with json parsing
// is is generated by getFacts and received by Facter
type stringmap map[string]interface{}

// cmdConfig stores the path to executable and defaults options to get unserializable response
type cmdConfig struct {
	// where to find facter executable
	// tip: somehow golang is able to find executable within PATH, so just a name is enough
	//		even if facter is just a ruby script
	executable string
	// default options passed to facter to get json output
	// overriden during tests to get non parseable output
	options []string
}

// newDefaultCmdConfig returns default working configuration "facter --json"
func newDefaultCmdConfig() *cmdConfig {
	return &cmdConfig{
		executable: "facter",
		options:    []string{"--json"},
	}
}

// get facts from facter (external command)
// returns all keys if none requested
// if cmdOptions is nil, use defaults "facter --json"
func getFacts(
	keys []string,
	facterTimeout time.Duration,
	cmdConfig *cmdConfig,
) (*stringmap, *time.Time, error) {

	// nil means use defaults
	if cmdConfig == nil {
		cmdConfig = newDefaultCmdConfig()
	}

	// copy given options into args
	// > args := make([]string, len(cmdConfig.options))
	// > copy(args, cmdConfig.options)
	args := append([]string{}, cmdConfig.options...)
	args = append(args, keys...)

	// communication channels
	jobCompletedChan := make(chan struct{})
	timeoutChan := time.After(facterTimeout)

	// closure to get data from command
	var err error
	output := make([]byte, 0, 1024)

	// execute command and capture the output
	go func() {
		output, err = exec.Command(cmdConfig.executable, args...).Output()
		jobCompletedChan <- struct{}{}
	}()

	// wait for done
	select {
	case <-timeoutChan:
		// timeout
		return nil, nil, errors.New("Facter plugin: fact gathering timeout")
	case <-jobCompletedChan:
		// success
	}

	if err != nil {
		log.Printf("Exec failed: %s\n", err)
		return nil, nil, err
	}

	// remember time of execution ended
	timestamp := time.Now()

	// parse output
	var data stringmap
	err = json.Unmarshal(output, &data)
	if err != nil {
		log.Printf("Unmarshal failed: %s\n", err)
		return nil, nil, err
	}
	return &data, &timestamp, nil
}
