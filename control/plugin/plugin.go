package plugin

// WARNING! Do not import "fmt" and print from a plugin to stdout!
import (
	"crypto/rsa"
	"time"
)

var (
	// Timeout settings
	// How much time must elapse before a lack of Ping results in a timeout
	PingTimeoutDurationDefault = time.Millisecond * 1500
	// How many succesive PingTimeouts must occur to equal a failure.
	PingTimeoutLimit = 3
)

const (
	// List of plugin type
	CollectorPluginType PluginType = iota
	PublisherPluginType
	ProcessorPluginType
)

const (
	// List of content types
	PulseGobContentType ContentType = iota
	PulseJsonContentType
)

const (
	// List of plugin response states
	PluginSuccess PluginResponseState = iota
	PluginFailure
)

var (
	// Array matching plugin type enum to a string
	// note: in string represenation we use lower case
	types = [...]string{
		"collector",
		"publisher",
		"processor",
	}
)

var (
	contentTypes = [...]string{
		"pulse.gob",
		"pulse.json",
	}
)

type PluginResponseState int

type PluginType int

type ContentType int

// Plugin interface
type Plugin interface {
}

// Returns string for matching enum plugin type
func (p PluginType) String() string {
	return types[p]
}

// Arguments passed to startup of Plugin
type Arg struct {
	// Plugin file path to binary
	PluginLogPath string
	// A public key from control used to verify RPC calls - not implemented yet
	ControlPubKey *rsa.PublicKey
	// Ping timeout duration
	PingTimeoutDuration time.Duration

	NoDaemon bool
	// The listen port
	listenPort string
}

func NewArg(pubkey *rsa.PublicKey, logpath string) Arg {
	return Arg{
		ControlPubKey:       pubkey,
		PluginLogPath:       logpath,
		PingTimeoutDuration: PingTimeoutDurationDefault,
	}
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

// // ConfigPolicy for plugin
// type ConfigPolicy struct {
// }

// PluginMeta for plugin
type PluginMeta struct {
	Name    string
	Version int
	Type    PluginType
}

// NewPluginMeta constructs and returns a PluginMeta struct
func NewPluginMeta(name string, version int, pluginType PluginType) *PluginMeta {
	return &PluginMeta{
		Name:    name,
		Version: version,
		Type:    pluginType,
	}
}

// Start starts a plugin
func Start(m *PluginMeta, c Plugin, requestString string) (error, int) {
	sessionState, sErr, retCode := NewSessionState(requestString)
	if sErr != nil {
		return sErr, retCode
	}

	//should we be initializing this
	sessionState.LastPing = time.Now()

	switch m.Type {
	case CollectorPluginType:
		r := &Response{
			Type:  CollectorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		err, retCode := StartCollector(c.(CollectorPlugin), sessionState, r)
		if err != nil {
			return sErr, retCode
		}
	case PublisherPluginType:
		r := &Response{
			Type:  PublisherPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		err, retCode := StartPublisher(c.(PublisherPlugin), sessionState, r)
		if err != nil {
			return sErr, retCode
		}
	}

	return nil, retCode
}
