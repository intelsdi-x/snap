package plugin

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"time"
)

// A plugin that is executable as a forked process on *Linux.
type ExecutablePlugin struct {
	cmd    *exec.Cmd
	stdout io.Reader
}

// A interface representing an executable plugin.
type PluginExecutor interface {
	Kill() error
	Wait() error
	ResponseReader() io.Reader
}

type ExecutablePluginController interface {
	GenerateArgs(bool) Arg
}

// Starts the plugin and returns error if one ocurred. This is non blocking.
func (e *ExecutablePlugin) Start() error {
	return e.cmd.Start()
}

// Kills the plugin and returns error if one ocurred. This is blocking.
func (e *ExecutablePlugin) Kill() error {
	return e.cmd.Process.Kill()
}

// Waits for plugin to halt. If error is returned then plugin stopped with error. If not plugin stopped safely.
func (e *ExecutablePlugin) Wait() error {
	return e.cmd.Wait()
}

// The STDOUT pipe for the plugin as io.Reader. Use to read from plugin process STDOUT.
func (e *ExecutablePlugin) ResponseReader() io.Reader {
	return e.stdout
}

// Initialize a new ExecutablePlugin from path to executable and daemon mode (true or false)
func NewExecutablePlugin(c ExecutablePluginController, path string, daemon bool) (*ExecutablePlugin, error) {
	jsonArgs, err := json.Marshal(c.GenerateArgs(daemon))
	if err != nil {
		return nil, err
	}
	// Init the cmd
	cmd := new(exec.Cmd)
	cmd.Path = path
	cmd.Args = []string{path, string(jsonArgs)}
	// Link the stdout for response reading
	stdout, err2 := cmd.StdoutPipe()
	if err2 != nil {
		return nil, err2
	}
	// Init the ExecutablePlugin and return
	ePlugin := new(ExecutablePlugin)
	ePlugin.cmd = cmd
	ePlugin.stdout = stdout

	return ePlugin, nil
}

// Wait for response from started ExecutablePlugin. Returns Response or error.
func WaitForResponse(p PluginExecutor, timeout time.Duration) (*Response, error) {
	// The response we want to return

	var resp *Response = new(Response)
	var timeoutErr error
	var jsonErr error

	// Kill on timeout
	go func() {
		time.Sleep(timeout)
		timeoutErr = errors.New("Timeout waiting for response")
		p.Kill()
		return
	}()

	// Wait for response from ResponseReader
	scanner := bufio.NewScanner(p.ResponseReader())
	go func() {
		for scanner.Scan() {
			// Get bytes
			b := scanner.Bytes()
			// attempt to unmarshall into struct
			err := json.Unmarshal(b, resp)
			if err != nil {
				jsonErr = errors.New("JSONError - " + err.Error())
				return
			}
		}
	}()

	// Wait for PluginExecutor to respond
	err := p.Wait()
	// Return top level error
	if jsonErr != nil {
		return nil, jsonErr
	}
	// Return top level error
	if timeoutErr != nil {
		return nil, timeoutErr
	}
	// Return pExecutor.Wait() error
	if err != nil {
		// log.Printf("[CONTROL] Plugin stopped with error [%v]\n", err)
		return nil, err
	}
	// Return response
	return resp, nil
}
