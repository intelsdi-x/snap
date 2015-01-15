package control

import (
	"encoding/json"
	"io"
	"os/exec"

	"github.com/intelsdilabs/pulse/control/plugin"
)

// A plugin that is executable as a forked process on *Linux.
type ExecutablePlugin struct {
	cmd    *exec.Cmd
	stdout io.Reader
}

type ExecutablePluginController interface {
	GenerateArgs(bool) plugin.Arg
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
func newExecutablePlugin(c ExecutablePluginController, path string, daemon bool) (*ExecutablePlugin, error) {
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
