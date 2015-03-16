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
// is is created by getFacts from actual values received from system
// maps a fact name to some untype value, received from Facter
type facts map[string]interface{}

// cmdConfig stores the path to executable and defaults options to get unserializable response
type cmdConfig struct {
	// where to find facter executable
	// tip: somehow golang is able to find executable within PATH, so just a name is enough
	//		even if facter is just a ruby script
	executable string
	// default options passed to facter to get json output
	// overriden during tests to get non parseable output
	// by options I mean flags as -h -v or --json (not positional arguments like foo bar)
	options []string
}

// newDefaultCmdConfig returns default working configuration "facter --json"
func newDefaultCmdConfig() *cmdConfig {
	return &cmdConfig{
		executable: "facter",
		options:    []string{"--json"},
	}
}

// get facts from facter (external command) requsted by names
// returns all facts if none requested
// if cmdOptions is nil, use defaults "facter --json"
func getFacts(
	names []string,
	facterTimeout time.Duration,
	cmdConfig *cmdConfig,
) (*facts, *time.Time, error) {

	// nil means use defaults
	if cmdConfig == nil {
		cmdConfig = newDefaultCmdConfig()
	}

	// copy given options into args
	// > args := make([]string, len(cmdConfig.options))
	// > copy(args, cmdConfig.options)
	args := append([]string{}, cmdConfig.options...)
	args = append(args, names...)

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

	facts := make(facts)

	// parse output
	err = json.Unmarshal(output, &facts)
	if err != nil {
		log.Printf("Unmarshal failed: %s\n", err)
		return nil, nil, err
	}
	return &facts, &timestamp, nil
}
