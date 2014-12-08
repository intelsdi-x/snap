package control

import (
	"io"
	"os/exec"
)

// A plugin that is executable as a forked process on *Linux.
type ExecutablePlugin struct {
	cmd    *exec.Cmd
	stdout io.Reader
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
