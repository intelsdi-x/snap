package control

import (
	"errors"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

type PluginExecutor interface {
	Kill() error
	Wait() error
}

func WaitForPluginResponse(pExecutor PluginExecutor, timeout time.Duration) (*plugin.Response, error) {

	// The response we want to return

	var resp *plugin.Response = new(plugin.Response)
	var topErr error

	// Kill on timeout
	go func() {
		time.Sleep(timeout)
		topErr = errors.New("Timeout waiting for response")
		pExecutor.Kill()
		return
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
	// Return top level error
	if topErr != nil {
		return nil, topErr
	}
	// Return pExecutor.Wait() error
	if err != nil {
		// log.Printf("[CONTROL] Plugin stopped with error [%v]\n", err)
		return nil, err
	}
	// Return response
	return resp, nil
}
