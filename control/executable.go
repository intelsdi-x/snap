package control

import (
	"encoding/json"
	"io"
	"os/exec"
)

type ExecutablePlugin struct {
	Cmd    *exec.Cmd
	Stdout io.Reader
}

// Starts the plugin and returns error or nil. This is non blocking.
func (e *ExecutablePlugin) Start() error {
	return e.Cmd.Start()
}

func (e *ExecutablePlugin) Kill() error {
	return e.Cmd.Process.Kill()
}

func (e *ExecutablePlugin) Wait() error {
	return e.Cmd.Wait()
}

func (e *ExecutablePlugin) StdoutPipe() io.Reader {
	return e.Stdout
}

// Take the path and daemon mode and returns *ExecutablePlugin
func (p *pluginControl) NewExecutablePlugin(path string, daemon bool) (*ExecutablePlugin, error) {
	jsonArgs, err := json.Marshal(p.GenerateArgs(daemon))
	if err != nil {
		return nil, err
	}

	cmd := new(exec.Cmd)
	cmd.Path = path
	cmd.Args = []string{path, string(jsonArgs)}

	stdout, err2 := cmd.StdoutPipe()
	if err2 != nil {
		return nil, err2
	}

	ePlugin := new(ExecutablePlugin)
	ePlugin.Cmd = cmd
	ePlugin.Stdout = stdout

	return ePlugin, nil
}
