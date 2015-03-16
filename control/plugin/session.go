package plugin

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

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
	ResetHeartbeat()

	generateResponse(r *Response) []byte
	heartbeatWatch(killChan chan int)
	isDaemon() bool
}

// Arguments passed to ping
type PingArgs struct{}

type KillArgs struct {
	Reason string
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

// Ping returns nothing in normal operation
func (s *SessionState) Ping(arg PingArgs, b *bool) error {
	// For now we return nil. We can return an error if we are shutting
	// down or otherwise in a state we should signal poor health.
	// Reply should contain any context.
	s.ResetHeartbeat()
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

func (s *SessionState) ResetHeartbeat() {
	s.LastPing = time.Now()
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
