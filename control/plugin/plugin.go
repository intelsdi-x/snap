package plugin

// Config Policy
// task > control > default

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"time"
)

const (
	// Timeout settings
	// How much time must elapse before a lack of Ping results in a timeout
	PingTimeoutDuration = time.Second * 5
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

type PluginResponseState int

type Plugin interface{}

type ConfigPolicy struct{}

type PluginMeta struct {
	Name    string
	Version int
}

type MetricType struct {
	Namespace               []string
	LastAdvertisedTimestamp int64
}

func NewMetricType(ns []string) *MetricType {
	return &MetricType{
		Namespace:               ns,
		LastAdvertisedTimestamp: time.Now().Unix(),
	}
}

type PluginType int

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
	// The listen port requested - optional, defaults to 0 via InitSessionState()
	ListenPort string
	// Whether to run as daemon to exit after sending response
	RunAsDaemon bool
}

// Started plugin session state
type SessionState struct {
	*Arg
	Token         string
	ListenAddress string
	LastPing      time.Time
	Logger        *log.Logger
}

func InitSessionState(path, pluginArgsMsg string) (*SessionState, error) {
	pluginArg := new(Arg)
	err := json.Unmarshal([]byte(pluginArgsMsg), pluginArg)
	if err != nil {
		return nil, err
	}

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

	return &SessionState{Arg: pluginArg, Token: rs}, nil
}

// Arguments passed to ping
type PingArgs struct{}

func (s *SessionState) Ping(arg PingArgs, b *bool) error {
	// For now we return nil. We can return an error if we are shutting
	// down or otherwise in a state we should signal poor health.
	// Reply should contain any context.
	s.LastPing = time.Now()
	s.Logger.Println("Ping received")
	return nil
}

// Arguments passed to Kill
type KillArgs struct {
	Reason string
}

func (s *SessionState) Kill(arg KillArgs, b *bool) error {
	// Right now we have no coordination needed. In the future we should
	// add control to wait on a lock before halting.
	s.Logger.Printf("Kill called by agent, reason: %s\n", arg.Reason)
	go func() {
		time.Sleep(time.Second * 2)
		s.haltPlugin(3)
	}()
	return nil
}

// Response from started plugin
type Response struct {
	Meta              PluginMeta
	ListenAddress     string
	Token             string
	Type              PluginType
	CollectorResponse GetMetricTypesReply

	// State is a signal from plugin to control that it passed
	// its own loading requirements
	State        PluginResponseState
	ErrorMessage string
}

func (s *SessionState) GenerateResponse(r Response) []byte {
	// Add common plugin response properties
	r.ListenAddress = s.ListenAddress
	r.Token = s.Token
	rs, _ := json.Marshal(r)
	return rs
}

func (s *SessionState) haltPlugin(code int) {
	s.Logger.Printf("Halting with exit code (%d)\n", code)
	os.Exit(code)
}

func (s *SessionState) heartbeatWatch(killChan chan (struct{})) {
	s.Logger.Println("Heartbeat started")
	count := 0
	for {
		if time.Now().Sub(s.LastPing) >= PingTimeoutDuration {
			count++
			if count >= PingTimeoutLimit {
				s.Logger.Println("Heartbeat timeout expired")
				defer close(killChan)
				return
			}
		} else {
			s.Logger.Println("Heartbeat timeout reset")
			// Reset count
			count = 0
		}
		time.Sleep(PingTimeoutDuration)
	}
}
