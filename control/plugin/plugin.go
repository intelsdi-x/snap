package plugin

// WARNING! Do not import "fmt" and print from a plugin to stdout!
import (
	"crypto/rsa"
	"fmt" // Don't use "fmt.Print*"
	"log"
	"regexp"
	"runtime"
	"time"
)

// Plugin type
type PluginType int

// Returns string for matching enum plugin type
func (p PluginType) String() string {
	return types[p]
}

const (
	CollectorPluginType PluginType = iota
	PublisherPluginType
	ProcessorPluginType
)

// Plugin response states
type PluginResponseState int

const (
	PluginSuccess PluginResponseState = iota
	PluginFailure
)

var (
	// Timeout settings
	// How much time must elapse before a lack of Ping results in a timeout
	PingTimeoutDurationDefault = time.Millisecond * 1500
	// How many succesive PingTimeouts must occur to equal a failure.
	PingTimeoutLimit = 3

	// Array matching plugin type enum to a string
	// note: in string represenation we use lower case
	types = [...]string{
		"collector",
		"publisher",
		"processor",
	}
)

type Plugin interface {
}

// PluginMeta for plugin
type PluginMeta struct {
	Name    string
	Version int
	Type    PluginType
	// Content types accepted by this plugin in priority order
	// pulse.* means any pulse type
	AcceptedContentTypes []string
	// Return content types in priority order
	// This is only really valid on processors
	ReturnedContentTypes []string
}

// NewPluginMeta constructs and returns a PluginMeta struct
func NewPluginMeta(name string, version int, pluginType PluginType, acceptContentTypes, returnContentTypes []string) *PluginMeta {
	// An empty accepted content type default to "pulse.*"
	if len(acceptContentTypes) == 0 {
		acceptContentTypes = append(acceptContentTypes, "pulse.*")
	}
	// Validate content type formats
	for _, s := range acceptContentTypes {
		b, e := regexp.MatchString(`^[a-z0-9*]+\.[a-z0-9*]+$`, s)
		if e != nil {
			panic(e)
		}
		if !b {
			panic(fmt.Sprintf("Bad accept content type [%s] for [%d] [%s]", name, version, s))
		}
	}
	for _, s := range returnContentTypes {
		b, e := regexp.MatchString(`^[a-z0-9*]+\.[a-z0-9*]+$`, s)
		if e != nil {
			panic(e)
		}
		if !b {
			panic(fmt.Sprintf("Bad return content type [%s] for [%d] [%s]", name, version, s))
		}
	}

	return &PluginMeta{
		Name:                 name,
		Version:              version,
		Type:                 pluginType,
		AcceptedContentTypes: acceptContentTypes,
		ReturnedContentTypes: returnContentTypes,
	}
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

// Start starts a plugin where:
// PluginMeta - base information about plugin
// Plugin - either CollectorPlugin or PublisherPlugin
// requestString - plugins arguments (marshaled json of control/plugin Arg struct)
// returns an error and exitCode (exitCode from SessionState initilization or plugin termination code)
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
	case ProcessorPluginType:
		r := &Response{
			Type:  ProcessorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		err, retCode := StartProcessor(c.(ProcessorPlugin), sessionState, r)
		if err != nil {
			return sErr, retCode
		}
	}

	return nil, retCode
}

func catchPluginPanic(l *log.Logger) {
	if err := recover(); err != nil {
		trace := make([]byte, 4096)
		count := runtime.Stack(trace, true)
		l.Printf("Recover from panic: %s\n", err)
		l.Printf("Stack of %d bytes: %s\n", count, trace)
		panic(err)
	}
}
