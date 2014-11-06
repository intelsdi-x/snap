package plugin

// Config Policy
// task > control > default

import (
	"crypto/rsa"
	"encoding/json"
)

// Plugin interface
type Plugin interface {
}

// Started plugin session state
type SessionState struct {
	*Arg
}

// Arguments passed to startup of Plugin
type Arg struct {
	// Plugin file path to binary
	PluginLogPath string
	// A public key from control used to verify RPC calls - not implemented yet
	ControlPubKey *rsa.PublicKey
	// The listen port requested - optional, defaults to 0 via InitSessionState()
	ListenPort string
}

// Response from started plugin
type Response struct {
	ListenAddress string
	Token         string
}

type ConfigPolicy struct {
}

type PluginMeta struct {
	Name    string
	Version int
}

func (p *PluginMeta) Status(a string, b *string) error {
	return nil
}

func InitSessionState(path string, pluginArgsMsg string) *SessionState {
	pluginArg := new(Arg)
	json.Unmarshal([]byte(pluginArgsMsg), pluginArg)

	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	if pluginArg.ListenPort == "" {
		pluginArg.ListenPort = "0"
	}

	return &SessionState{Arg: pluginArg}
}
