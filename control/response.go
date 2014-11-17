package control

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

type PluginExecutor interface {
	Kill() error
	Wait() error
	StdoutPipe() io.Reader
}

func WaitForPluginResponse(pExecutor PluginExecutor, timeout time.Duration) (*plugin.Response, error) {

	// The response we want to return

	var resp *plugin.Response = new(plugin.Response)
	var timeoutErr error
	var jsonErr error

	// Kill on timeout
	go func() {
		time.Sleep(timeout)
		timeoutErr = errors.New("Timeout waiting for response")
		pExecutor.Kill()
		return
	}()

	// Wait for response
	scanner := bufio.NewScanner(pExecutor.StdoutPipe())
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

	err := pExecutor.Wait()
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
