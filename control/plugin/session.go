/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugin

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encoding"
	"github.com/intelsdi-x/snap/control/plugin/encrypter"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
)

// Session interface
type Session interface {
	Ping([]byte, *[]byte) error
	Kill([]byte, *[]byte) error
	GetConfigPolicy([]byte, *[]byte) error
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

	SetKey(SetKeyArgs, *[]byte) error
	setKey([]byte)

	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error

	DecryptKey([]byte) ([]byte, error)
}

// Arguments passed to ping
type PingArgs struct{}

type KillArgs struct {
	Reason string
}

// Started plugin session state
type SessionState struct {
	*Arg
	*encrypter.Encrypter
	encoding.Encoder

	LastPing time.Time

	plugin        Plugin
	token         string
	listenAddress string
	killChan      chan int
	logger        *log.Logger
	privateKey    *rsa.PrivateKey
	encoder       encoding.Encoder
}

type GetConfigPolicyArgs struct{}

type GetConfigPolicyReply struct {
	Policy *cpolicy.ConfigPolicy
}

// GetConfigPolicy returns the plugin's policy
func (s *SessionState) GetConfigPolicy(args []byte, reply *[]byte) error {
	defer catchPluginPanic(s.Logger())

	s.logger.Println("GetConfigPolicy called")

	policy, err := s.plugin.GetConfigPolicy()
	if err != nil {
		return errors.New(fmt.Sprintf("GetConfigPolicy call error : %s", err.Error()))
	}

	r := GetConfigPolicyReply{Policy: policy}
	*reply, err = s.Encode(r)
	if err != nil {
		return err
	}

	return nil
}

// Ping returns nothing in normal operation
func (s *SessionState) Ping(arg []byte, reply *[]byte) error {
	// For now we return nil. We can return an error if we are shutting
	// down or otherwise in a state we should signal poor health.
	// Reply should contain any context.
	s.ResetHeartbeat()
	s.logger.Println("Ping received")
	*reply = []byte{}
	return nil
}

// Kill will stop a running plugin
func (s *SessionState) Kill(args []byte, reply *[]byte) error {
	a := &KillArgs{}
	err := s.Decode(args, a)
	if err != nil {
		return err
	}
	s.logger.Printf("Kill called by agent, reason: %s\n", a.Reason)
	go func() {
		time.Sleep(time.Second * 2)
		s.killChan <- 0
	}()
	*reply = []byte{}
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

type SetKeyArgs struct {
	Key []byte
}

func (s *SessionState) SetKey(args SetKeyArgs, reply *[]byte) error {
	s.logger.Println("SetKey called")
	out, err := s.DecryptKey(args.Key)
	if err != nil {
		return err
	}
	s.Key = out
	return nil
}

func (s *SessionState) setKey(key []byte) {
	s.Key = key
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
		if time.Since(s.LastPing) >= s.PingTimeoutDuration {
			count++
			s.logger.Printf("Heartbeat timeout %v of %v.  (Duration between checks %v)", count, PingTimeoutLimit, s.PingTimeoutDuration)
			if count >= PingTimeoutLimit {
				s.logger.Println("Heartbeat timeout expired!")
				defer close(killChan)
				return
			}
		} else {
			s.logger.Println("Heartbeat timeout reset")
			// Reset count
			count = 0
		}
		time.Sleep(s.PingTimeoutDuration)
	}
}

// NewSessionState takes the plugin args and returns a SessionState
// returns State or error and returnCode:
// 0 - ok
// 2 - error when unmarshaling pluginArgs
// 3 - cannot open error files
func NewSessionState(pluginArgsMsg string, plugin Plugin, meta *PluginMeta) (*SessionState, error, int) {
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

	// If no PingTimeoutDuration was provided we need to set it
	if pluginArg.PingTimeoutDuration == 0 {
		pluginArg.PingTimeoutDuration = PingTimeoutDurationDefault
	}

	// Generate random token for this session
	rb := make([]byte, 32)
	rand.Read(rb)
	rs := base64.URLEncoding.EncodeToString(rb)

	// Initialize a logger based on PluginLogPath
	truncOrAppend := os.O_TRUNC // truncate log file explicitly given by user
	// Empty or /tmp means use default tmp log (needs to be removed post-aAtruncOrAppendpha)
	if pluginArg.PluginLogPath == "" || pluginArg.PluginLogPath == "/tmp" {
		if runtime.GOOS == "windows" {
			pluginArg.PluginLogPath = `c:\TEMP\snap_plugin.log`
		} else {
			pluginArg.PluginLogPath = "/tmp/snap_plugin.log"
		}
		truncOrAppend = os.O_APPEND
	}
	lf, err := os.OpenFile(pluginArg.PluginLogPath, os.O_WRONLY|os.O_CREATE|truncOrAppend, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error opening log file: %v", err)), 3
	}
	logger := log.New(lf, ">>>", log.Ldate|log.Ltime)

	var enc encoding.Encoder
	switch meta.RPCType {
	case JSONRPC:
		enc = encoding.NewJsonEncoder()
	case NativeRPC:
		enc = encoding.NewGobEncoder()
	case GRPC:
		enc = encoding.NewGobEncoder()
		//TODO(CDR): re-think once content-types is settled
	}
	ss := &SessionState{
		Arg:     pluginArg,
		Encoder: enc,

		plugin:   plugin,
		token:    rs,
		killChan: make(chan int),
		logger:   logger,
	}

	if !meta.Unsecure {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err, 2
		}
		encrypt := encrypter.New(nil, key)
		enc.SetEncrypter(encrypt)
		ss.Encrypter = encrypt
		ss.privateKey = key
	}
	return ss, nil, 0
}

func init() {
	gob.RegisterName("conf_value_string", *(&ctypes.ConfigValueStr{}))
	gob.RegisterName("conf_value_int", *(&ctypes.ConfigValueInt{}))
	gob.RegisterName("conf_value_float", *(&ctypes.ConfigValueFloat{}))
	gob.RegisterName("conf_value_bool", *(&ctypes.ConfigValueBool{}))

	gob.RegisterName("conf_policy_node", cpolicy.NewPolicyNode())
	gob.RegisterName("conf_data_node", &cdata.ConfigDataNode{})
	gob.RegisterName("conf_policy_string", &cpolicy.StringRule{})
	gob.RegisterName("conf_policy_int", &cpolicy.IntRule{})
	gob.RegisterName("conf_policy_float", &cpolicy.FloatRule{})
	gob.RegisterName("conf_policy_bool", &cpolicy.BoolRule{})
}
