package plugin

// Config Policy
// task > control > default

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
)

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	PublisherPluginType
	ProcesserPluginType
)

const (
	// List of plugin response states
	PluginSuccess PluginResponseState = iota
	PluginFailure
)

var (
	// Slice matching plugin type enum to a string
	// note: in string represenation we use lower case
	types = [...]string{
		"collector",
		"publisher",
		"processor",
	}
)

type PluginResponseState int

type PluginType int

// Plugin interface
type Plugin interface {
}

// Returns string for matching enum plugin type
func (p PluginType) String() string {
	return types[p]
}

// Started plugin session state
type SessionState struct {
	*Arg
	Token         string
	ListenAddress string
}

// Arguments passed to startup of Plugin
type Arg struct {
	// Plugin file path to binary
	PluginLogPath string
	// A public key from control used to verify RPC calls - not implemented yet
	ControlPubKey *rsa.PublicKey
	// The listen port requested - optional, defaults to 0 via InitSessionState()
	ListenPort string
	// Whether to run as daemon to exit after sending response
	RunAsDaemon bool
}

// Response from started plugin
type Response struct {
	Meta          PluginMeta
	ListenAddress string
	Token         string
	Type          PluginType
	// State is a signal from plugin to control that it passed
	// its own loading requirements
	State        PluginResponseState
	ErrorMessage string
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

func (s *SessionState) GenerateResponse(r Response) []byte {
	// Add common plugin response properties
	r.ListenAddress = s.ListenAddress
	r.Token = s.Token
	rs, _ := json.Marshal(r)
	return rs
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

	// Generate random token for this session
	rb := make([]byte, 32)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)

	return &SessionState{Arg: pluginArg, Token: rs}
}
