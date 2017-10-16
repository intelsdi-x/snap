/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var execLogger = log.WithField("_module", "plugin-exec")

type ExecutablePlugin struct {
	name   string
	cmd    command
	stdout io.Reader
	stderr io.Reader
}

// An interface for the interactions ExecutablePlugin has with an exec.Cmd
// This way, the underlying Cmd can be mocked.
type command interface {
	Start() error
	Kill() error
	Path() string
}

// The implementation of command used here.
type commandWrapper struct {
	cmd *exec.Cmd
}

func (cw *commandWrapper) Path() string { return cw.cmd.Path }
func (cw *commandWrapper) Kill() error {
	// first, kill the process wrapped up in the commandWrapper
	if cw.cmd.Process == nil {
		err := fmt.Errorf("Process for plugin '%s' not started; cannot kill", path.Base(cw.Path()))
		log.WithFields(log.Fields{
			"_block": "Kill",
		}).Warn(err)
		return err
	} else if err := cw.cmd.Process.Kill(); err != nil {
		log.WithFields(log.Fields{
			"_block": "Kill",
		}).Error(err)
		return err
	}
	// then wait for it to exit (so that we don't have any zombie processes kicking
	// around the system)
	_, err := cw.cmd.Process.Wait()
	return err
}
func (cw *commandWrapper) Start() error { return cw.cmd.Start() }

// NewExecutablePlugin returns a new ExecutablePlugin.
func NewExecutablePlugin(a Arg, commands ...string) (*ExecutablePlugin, error) {
	jsonArgs, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	cmd := &exec.Cmd{
		Path: commands[0],
		Args: append(commands, string(jsonArgs)),
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	return &ExecutablePlugin{
		cmd:    &commandWrapper{cmd},
		stdout: stdout,
		stderr: stderr,
	}, nil
}

// Run executes the plugin and waits for a response, or times out.
func (e *ExecutablePlugin) Run(timeout time.Duration) (Response, error) {
	var (
		respReceived bool
		resp         Response
		err          error
		respBytes    []byte
	)

	doneChan := make(chan struct{})
	stdOutScanner := bufio.NewScanner(e.stdout)

	// Start the command and begin reading its output.
	if err = e.cmd.Start(); err != nil {
		return resp, err
	}

	e.captureStderr()
	go func() {
		for {
			for stdOutScanner.Scan() {
				// The first chunk from the scanner is the plugin's response to the
				// handshake.  Once we've received that, we can begin to forward
				// logs on to snapteld's log.
				if !respReceived {
					respBytes = stdOutScanner.Bytes()
					err = json.Unmarshal(respBytes, &resp)
					respReceived = true
					close(doneChan)
				} else {
					execLogger.WithFields(log.Fields{
						"plugin": e.name,
						"io":     "stdout",
					}).Debug(stdOutScanner.Text())
				}
			}

			if errScanner := stdOutScanner.Err(); errScanner != nil {
				reader := bufio.NewReader(e.stdout)
				log, errRead := reader.ReadString('\n')
				if errRead == io.EOF {
					break
				}

				execLogger.
					WithField("plugin", path.Base(e.cmd.Path())).
					WithField("io", "stdout").
					WithField("scanner_err", errScanner).
					WithField("read_string_err", errRead).
					Warn(log)

				continue //scanner finished with errors so try to scan once again
			}
			break //scanner finished scanning without errors so break the loop
		}
	}()

	// Wait until:
	//   a) We receive a signal that the plugin has responded
	// OR
	//   b) The timeout expires
	select {
	case <-doneChan:
	case <-time.After(timeout):
		// We timed out waiting for the plugin's response.  Set err.
		err = fmt.Errorf("timed out waiting for plugin %s", path.Base(e.cmd.Path()))
	}
	if err != nil {
		execLogger.WithFields(log.Fields{
			"received_response": string(respBytes),
		}).Error("error loading plugin")

		// Kill the plugin if we failed to load it.
		e.Kill()
	}
	lowerName := strings.ToLower(resp.Meta.Name)
	if lowerName != resp.Meta.Name {
		execLogger.WithFields(log.Fields{
			"plugin-name":    resp.Meta.Name,
			"plugin-version": resp.Meta.Version,
			"plugin-type":    resp.Type.String(),
		}).Warning("uppercase plugin name")
	}
	resp.Meta.Name = lowerName

	return resp, err
}

func (e *ExecutablePlugin) SetName(name string) {
	e.name = name
}

func (e *ExecutablePlugin) Kill() error {
	return e.cmd.Kill()
}

func (e *ExecutablePlugin) captureStderr() {
	stdErrScanner := bufio.NewScanner(e.stderr)
	go func() {
		for {
			for stdErrScanner.Scan() {
				execLogger.
					WithField("plugin", e.name).
					WithField("io", "stderr").
					Debug(stdErrScanner.Text())
			}

			if errScanner := stdErrScanner.Err(); errScanner != nil {
				reader := bufio.NewReader(e.stderr)
				log, errRead := reader.ReadString('\n')
				if errRead == io.EOF {
					break
				}

				execLogger.
					WithField("plugin", e.name).
					WithField("io", "stderr").
					WithField("scanner_err", errScanner).
					WithField("read_string_err", errRead).
					Warn(log)

				continue //scanner finished with errors so try to scan once again
			}
			break //scanner finished scanning without errors so break the loop
		}
	}()
}
