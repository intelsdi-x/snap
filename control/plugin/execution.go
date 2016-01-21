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
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	execLogger = logrus.WithField("_module", "control-plugin-execution")
)

const (
	// enums for different waiting states
	pluginKilled      waitSignal = iota // plugin was killed
	pluginTimeout                       // plugin timed out
	pluginResponseOk                    // plugin response received (valid)
	pluginResponseBad                   // plugin response received (invalid)
)

// A plugin that is executable as a forked process on *Linux.
type ExecutablePlugin struct {
	cmd    *exec.Cmd
	stdout io.Reader
	stderr io.Reader
	args   Arg
}

// A interface representing an executable plugin.
type pluginExecutor interface {
	Kill() error
	WaitForExit() error
	ResponseReader() io.Reader
	ErrorResponseReader() io.Reader
}

type waitSignal int

type waitSignalValue struct {
	Signal   waitSignal
	Response *Response
	Error    *error
}

// Starts the plugin and returns error if one occurred. This is non blocking.
func (e *ExecutablePlugin) Start() error {
	err := e.cmd.Start()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"_module":  "control-executableplugin",
			"_block":   "start",
			"cmd path": e.cmd.Path,
			"cmd args": e.cmd.Args,
			"error":    err.Error(),
		}).Error("error in starting executable plugin")
	}
	return err
}

// Kills the plugin and returns error if one occurred. This is blocking.
func (e *ExecutablePlugin) Kill() error {
	execLogger.WithField("path", e.cmd.Path).Debug("Hard killing plugin")
	return e.cmd.Process.Kill()
}

// Waits for plugin to halt. If error is returned then plugin stopped with error. If not plugin stopped safely.
func (e *ExecutablePlugin) WaitForExit() error {
	return e.cmd.Wait()
}

// The STDOUT pipe for the plugin as io.Reader. Use to read from plugin process STDOUT.
func (e *ExecutablePlugin) ResponseReader() io.Reader {
	return e.stdout
}

// The STDERR pipe for the plugin as a io.reader
func (e *ExecutablePlugin) ErrorResponseReader() io.Reader {
	return e.stderr
}

// Initialize a new ExecutablePlugin from path to executable and daemon mode (true or false)
func NewExecutablePlugin(a Arg, path string) (*ExecutablePlugin, error) {
	jsonArgs, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	// Init the cmd
	cmd := new(exec.Cmd)
	cmd.Path = path
	cmd.Args = []string{path, string(jsonArgs)}
	// Link the stdout for response reading
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	// Init the ExecutablePlugin and return
	ePlugin := new(ExecutablePlugin)
	ePlugin.cmd = cmd
	ePlugin.stdout = stdout
	ePlugin.args = a
	ePlugin.stderr = stderr

	return ePlugin, nil
}

// Waits for a plugin response from a started plugin
func (e *ExecutablePlugin) WaitForResponse(timeout time.Duration) (*Response, error) {
	r, err := waitHandling(e, timeout, e.args.PluginLogPath)
	return r, err
}

// Private method which handles behavior for wait for response for daemon and non-daemon modes.
func waitHandling(p pluginExecutor, timeout time.Duration, logpath string) (*Response, error) {
	log := execLogger.WithField("_block", "waitHandling")

	/*
		Bit of complex behavior so some notes:
			A. We need to wait for three scenarios depending on the daemon setting
					1)	plugin is killed (like a safe exit in non-daemon)
						causing WaitForExit to fire
					2) 	plugin timeout fires calling Kill() and causing
						WaitForExit to fire
					3)	A response is returned before either 1 or 2 occur

				notes:
					*	In daemon mode (daemon == true) we want to wait until (1) or
						(2 then 1) or (3) occurs and stop waiting right after.
					*	In non-daemon mode (daemon == false) we want to return on (1)
						or (2 then 1) regardless of whether (3) occurs before or after.

			B. We will start three go routines to handle
					1)	waiting for timeout, on timeout we signal timeout and then
						kill plugin
					2)	wait for exit, also known as wait for kill, on kill we fire
						proper code to waitChannel
					3)	wait for response, on response we fire proper code to waitChannel

			C. The wait behavior loops collecting
					1)	timeout signal, this is used to mark exit by timeout
					2)	killed signal, signal the plugin has stopped - this exits
						the loop for all scenarios
					3)	response received, signal the plugin has responded - this exits
						the loop if daemon == true, otherwise waits for (2)
					4)	response received but corrupt
	*/

	// wait channel
	waitChannel := make(chan waitSignalValue, 3)

	// send timeout signal to our channel on timeout
	log.Debug("timeout chan start")
	go waitForPluginTimeout(timeout, p, waitChannel)

	// send response received signal to our channel on response
	log.Debug("response chan start")
	go waitForResponseFromPlugin(p.ResponseReader(), waitChannel, logpath)

	// log stderr from the plugin
	go logStdErr(p.ErrorResponseReader(), logpath)

	// send killed plugin signal to our channel on kill
	log.Debug("kill chan start")
	go waitForKilledPlugin(p, waitChannel)

	// flag to indicate a timeout occurred
	var timeoutFlag bool
	// error value indicating a bad response was found
	var errResponse *error
	// var holding a good response (or nil if none was returned)
	var response *Response
	// Loop to wait for signals and return
	for {
		w := <-waitChannel
		switch w.Signal {
		case pluginTimeout: // plugin timeout signal received
			log.Debug("plugin timeout signal received")
			// If timeout received after response we are ok with it and
			// don't need to flip the timeout flag.
			if response == nil {
				log.Debug("timeout flag set")
				// We got a timeout without getting a response
				// set the flag
				timeoutFlag = true
				// Kill the plugin.
				p.Kill()
				break
			}
			log.Debug("timeout flag ignored because of response")

		case pluginKilled: // plugin killed signal received
			log.Error("plugin kill signal received")
			// We check a few scenarios and return based on how things worked out to this point
			// 1) If a bad response was received we return signalling this with an error (fail)
			if errResponse != nil {
				log.Error("returning with error (bad response)")
				return nil, *errResponse
			}
			// 2) If a timeout occurred we return that as error (fail)
			if timeoutFlag {
				log.Error("returning with error (timeout)")
				return nil, errors.New("timeout waiting for response")
			}
			// 3) If a good response was returned we return that with no error (success)
			if response != nil {
				log.Error("returning with response (after wait for kill)")
				return response, nil
			}
			// 4) otherwise we return no response and an error that no response was received (fail)
			log.Error("returning with error (killed without response)")
			// The kill could have been without error so we check if ExitError was returned and return
			// our own if not.
			if *w.Error != nil {
				return nil, *w.Error
			} else {
				return nil, errors.New("plugin died without sending response")
			}

		case pluginResponseOk: // plugin response (valid) signal received
			log.Debug("plugin response (ok) received")
			// If in daemon mode we can return now (succes) since the plugin will continue to run
			// if not we let the loop continue (to wait for kill)
			response = w.Response
			return response, nil

		case pluginResponseBad: // plugin response (invalid) signal received
			log.Error("plugin response (bad) received")
			// A bad response is end of game in all scerarios and indictive of an unhealthy or unsupported plugin
			// We save the response bad error var (for handling later on plugin kill)
			errResponse = w.Error
		}
	}
}

func waitForPluginTimeout(timeout time.Duration, p pluginExecutor, waitChannel chan waitSignalValue) {
	// sleep for timeout duration
	time.Sleep(timeout)
	// Check if waitChannel is closed. If it is we exit now.
	// Send out timeout signal, waiting method will still wait for exit caused by p.Kill
	// Because this channel is shared this ensures that the resulting kill signals the channel after
	// the response has already queued across it.
	waitChannel <- waitSignalValue{Signal: pluginTimeout}
}

func waitForResponseFromPlugin(r io.Reader, waitChannel chan waitSignalValue, logpath string) {
	lp := strings.TrimSuffix(logpath, filepath.Ext(logpath))
	lf, _ := os.OpenFile(lp+".stdout", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer lf.Close()
	logger := log.New(lf, "", log.Ldate|log.Ltime)
	processedResponse := false
	scanner := bufio.NewScanner(r)
	resp := new(Response)
	// scan until we get a response or reader is closed
OK:
	for scanner.Scan() {
		if !processedResponse {
			// Get bytes
			b := scanner.Bytes()
			// attempt to unmarshall into struct
			err := json.Unmarshal(b, resp)
			if err != nil {
				log.Println("JSON error in response: " + err.Error())
				log.Printf("response: \"%s\"\n", string(b))
				e := errors.New("JSONError - " + err.Error())
				// send plugin response received but bad
				waitChannel <- waitSignalValue{Signal: pluginResponseBad, Error: &e}
				// exit function
				return
			}
			// send plugin response received (valid)
			waitChannel <- waitSignalValue{Signal: pluginResponseOk, Response: resp}
			processedResponse = true
		} else {
			logger.Println(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		if err == bufio.ErrTooLong {
			reader := bufio.NewReader(r)
			logger.Println(reader.ReadLine())
			goto OK
		}
		logger.Println(err)
	}
}

func logStdErr(r io.Reader, logpath string) {
	lp := strings.TrimSuffix(logpath, filepath.Ext(logpath))
	lf, _ := os.OpenFile(lp+".stderr", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer lf.Close()
	logger := log.New(lf, "", log.Ldate|log.Ltime)
	scanner := bufio.NewScanner(r)
OK:
	for scanner.Scan() {
		logger.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		if err == bufio.ErrTooLong {
			reader := bufio.NewReader(r)
			logger.Println(reader.ReadLine())
			goto OK
		}
		logger.Println(err)
	}
}

func waitForKilledPlugin(p pluginExecutor, waitChannel chan waitSignalValue) {
	// simply wait for this to return (blocking method)
	// TODO, refactor not to block. In daemon mode this would hang for the life of process.
	// ideally this should check if running or waitChannel closed and then exit on either.
	e := p.WaitForExit()
	time.Sleep(time.Millisecond * 100)
	// send signal
	waitChannel <- waitSignalValue{Signal: pluginKilled, Error: &e}
}
