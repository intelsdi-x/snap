// Test helper for testing plugins
package helper

import (
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
func BuildPlugin(pluginType, pluginName string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	bPath := strings.Replace(wd, path.Join("/", "plugin", pluginType, pluginName), buildScript, 1)
	sPath := strings.Replace(wd, path.Join("/", "plugin", pluginType, pluginName), "", 1)

	fmt.Println(bPath, sPath, pluginType, pluginName)
	c := exec.Command(bPath, sPath, pluginType, pluginName)

	o, e := c.Output()
	fmt.Println(string(o))
	return e
}
