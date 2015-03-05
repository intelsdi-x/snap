// Plugin helper for building plugins
package helper

import (
	"strings"
)

const SEPARATOR = "/"

func GetPluginName(namespace *[]string) string {
	return strings.Join(*namespace, SEPARATOR)
}
