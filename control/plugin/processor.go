package plugin

import (
	"encoding/gob"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// Processor plugin
type ProcessorPlugin interface {
	Plugin
	Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error)
	GetConfigPolicyNode() cpolicy.ConfigPolicyNode
}

func init() {
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
	gob.Register(&cpolicy.IntRule{})
	gob.Register(&cpolicy.FloatRule{})
}
