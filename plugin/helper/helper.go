// Test helper for testing plugins
package helper

import (
	// "fmt"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	buildScript = "/scripts/build.sh"
)

// Attempts to make the plugins before each test.
// unless PULSE_BUILD_HELPER_NO_REBUILD is set
func BuildPlugin(pluginType, pluginName string) error {

	if os.Getenv("PULSE_BUILD_HELPER_NO_REBUILD") != "" {
		fmt.Println("skiping build because PULSE_BUILD_HELPER_NO_REBUILD is set")
		return nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	bPath := strings.Replace(wd, path.Join("/", "plugin", pluginType, pluginName), buildScript, 1)
	sPath := strings.Replace(wd, path.Join("/", "plugin", pluginType, pluginName), "", 1)

	fmt.Println(bPath, sPath, pluginType, pluginName)
	c := exec.Command(bPath, sPath, pluginType, pluginName)

	_, e := c.Output()
	if err != nil {
		panic(err)
	}
	return e
}
