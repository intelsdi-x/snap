/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
) (facts, error) {

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
		return nil, errors.New("Facter plugin: fact gathering timeout")
	case <-jobCompletedChan:
		// success
	}

	if err != nil {
		log.Printf("Exec failed: %s\n", err)
		return nil, err
	}

	// place to unmarsha values received from Facter
	facts := make(facts)

	// parse output
	err = json.Unmarshal(output, &facts)
	if err != nil {
		log.Printf("Unmarshal failed: %s\n", err)
		return nil, err
	}
	return facts, nil
}
