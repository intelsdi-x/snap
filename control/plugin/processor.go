package plugin

import "github.com/intelsdi-x/pulse/core/ctypes"

// Processor plugin
type ProcessorPlugin interface {
	Plugin
	Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error)
}
