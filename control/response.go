package control

import (
	// "bufio"
	// "encoding/json"
	"errors"
	"github.com/intelsdilabs/pulse/control/plugin"
	// // "io"
	// "log"
	"time"
)

type PluginExecutor interface {
	Kill() error
	Wait() error
}

func WaitForPluginResponse(pExecutor PluginExecutor, timeout time.Duration) (*plugin.Response, error) {

	// The response we want to return

	var resp *plugin.Response = new(plugin.Response)
	var topErr error

	pExecutor.Kill()

	// Kill on timeout
	go func() {
		time.Sleep(timeout)
		// if !cmd.ProcessState.Exited() {
		err := pExecutor.Kill()
		if err != nil {
			// It is possible it died on its own or was killed somewhere else
			// we just want to log this and not panic.
			// log.Println(err.Error())
		} else {
			topErr = errors.New("Timeout waiting for response")
			return
		}
		// }
	}()

	// // Wait for response
	// scanner := bufio.NewScanner(stdout)
	// go func() {
	// 	for scanner.Scan() {
	// 		// Get bytes
	// 		b := scanner.Bytes()
	// 		fmt.Println(string(b))
	// 		// attempt to unmarshall into struct
	// 		err := json.Unmarshal(b, resp)
	// 		if err != nil {
	// 			// log.Printf("[CONTROL] Response from plugin is invalid [%s]\n", err)
	// 			topErr = errors.New("JSONError - " + err.Error())
	// 			return
	// 		}
	// 		log.Printf("[CONTROL] Response received from plugin\n")
	// 	}
	// }()

	err := pExecutor.Wait()
	if topErr != nil {
		// log.Printf("[CONTROL] Plugin stopped with error [%v]\n", topErr)
		return nil, topErr
	}

	if err != nil {
		// log.Printf("[CONTROL] Plugin stopped with error [%v]\n", err)
		return nil, err
	}

	// log.Printf("[CONTROL] Plugin run completed\n")
	return resp, nil
}
