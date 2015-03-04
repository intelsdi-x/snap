package plugin

// Config Policy
// task > control > default

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

var (
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

type MetricType struct {
	namespace          []string
	lastAdvertisedTime time.Time
}

func (m *MetricType) Namespace() []string {
	return m.namespace
}

func (m *MetricType) LastAdvertisedTime() time.Time {
	return m.lastAdvertisedTime
}

func NewMetricType(ns []string, last time.Time) *MetricType {
	return &MetricType{
		namespace:          ns,
		lastAdvertisedTime: last,
	}
}

type PluginResponseState int

type PluginType int

// Plugin interface
type Plugin interface {
}

// Returns string for matching enum plugin type
func (p PluginType) String() string {
	return types[p]
}

// Session interface
type Session interface {
	Ping(arg PingArgs, b *bool) error
	Kill(arg KillArgs, b *bool) error
	Logger() *log.Logger
	ListenAddress() string
	SetListenAddress(string)
	ListenPort() string
	Token() string
	KillChan() chan int

	generateResponse(r *Response) []byte
	heartbeatWatch(killChan chan int)
	isDaemon() bool
}

// Started plugin session state
type SessionState struct {
	*Arg
	LastPing time.Time

	token         string
	listenAddress string
	killChan      chan int
	logger        *log.Logger
}

// Arguments passed to startup of Plugin
type Arg struct {
	// Plugin file path to binary
	PluginLogPath string
	// A public key from control used to verify RPC calls - not implemented yet
	ControlPubKey *rsa.PublicKey

	NoDaemon bool
	// The listen port
	listenPort string
}

func NewArg(pubkey *rsa.PublicKey, logpath string) Arg {
	return Arg{
		ControlPubKey: pubkey,
		PluginLogPath: logpath,
	}
}

// Arguments passed to ping
type PingArgs struct{}

type KillArgs struct {
	Reason string
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

// ConfigPolicy for plugin
type ConfigPolicy struct {
}

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

// Ping returns nothing in normal operation
func (s *SessionState) Ping(arg PingArgs, b *bool) error {
	// For now we return nil. We can return an error if we are shutting
	// down or otherwise in a state we should signal poor health.
	// Reply should contain any context.
	s.LastPing = time.Now()
	s.logger.Println("Ping received")
	return nil
}

// Kill will stop a running plugin
func (s *SessionState) Kill(arg KillArgs, b *bool) error {
	// Right now we have no coordination needed. In the future we should
	// add control to wait on a lock before halting.
	s.logger.Printf("Kill called by agent, reason: %s\n", arg.Reason)
	go func() {
		time.Sleep(time.Second * 2)
		s.killChan <- 0
	}()
	return nil
}

// Logger gets the SessionState logger
func (s *SessionState) Logger() *log.Logger {
	return s.logger
}

// ListenAddress gets the SessionState listen address
func (s *SessionState) ListenAddress() string {
	return s.listenAddress
}

//ListenPort gets the SessionState listen port
func (s *SessionState) ListenPort() string {
	return s.listenPort
}

// SetListenAddress sets SessionState listen address
func (s *SessionState) SetListenAddress(a string) {
	s.listenAddress = a
}

// Token gets the SessionState token
func (s *SessionState) Token() string {
	return s.token
}

// KillChan gets the SessionState killchan
func (s *SessionState) KillChan() chan int {
	return s.killChan
}

func (s *SessionState) isDaemon() bool {
	return !s.NoDaemon
}

func (s *SessionState) generateResponse(r *Response) []byte {
	// Add common plugin response properties
	r.ListenAddress = s.listenAddress
	r.Token = s.token
	rs, _ := json.Marshal(r)
	return rs
}

func (s *SessionState) heartbeatWatch(killChan chan int) {
	s.logger.Println("Heartbeat started")
	count := 0
	for {
		if time.Now().Sub(s.LastPing) >= PingTimeoutDuration {
			count++
			if count >= PingTimeoutLimit {
				s.logger.Println("Heartbeat timeout expired")
				defer close(killChan)
				return
			}
		} else {
			s.logger.Println("Heartbeat timeout reset")
			// Reset count
			count = 0
		}
		time.Sleep(PingTimeoutDuration)
	}
}

// NewSessionState takes the plugin args and returns a SessionState
func NewSessionState(pluginArgsMsg string) (*SessionState, error, int) {
	pluginArg := &Arg{}
	err := json.Unmarshal([]byte(pluginArgsMsg), pluginArg)
	if err != nil {
		return nil, err, 2
	}

	// If no port was provided we let the OS select a port for us.
	// This is safe as address is returned in the Response and keep
	// alive prevents unattended plugins.
	if pluginArg.listenPort == "" {
		pluginArg.listenPort = "0"
	}

	// Generate random token for this session
	rb := make([]byte, 32)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)

	var logger *log.Logger
	switch lp := pluginArg.PluginLogPath; lp {
	case "", "/tmp":
		// Empty means use default tmp log (needs to be removed post-alpha)
		lf, err := os.OpenFile("/tmp/pulse_plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error opening log file: %v", err)), 3
		}
		logger = log.New(lf, ">>>", log.Ldate|log.Ltime)
	default:
		lf, err := os.OpenFile(lp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error opening log file: %v", err)), 3
		}
		logger = log.New(lf, ">>>", log.Ldate|log.Ltime)
	}

	return &SessionState{
		Arg:      pluginArg,
		token:    rs,
		killChan: make(chan int),
		logger:   logger}, nil, 0
}

// Start starts a plugin
func Start(m *PluginMeta, c Plugin, p *ConfigPolicy, requestString string) (error, int) {

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
	}

	return nil, retCode
}
