package plugin

import "github.com/intelsdi-x/pulse/core/ctypes"

// Publisher plugin
type PublisherPlugin interface {
	Plugin
	Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error
}
