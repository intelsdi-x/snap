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

// WARNING! Do not import "fmt" and print from a plugin to stdout!
import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io" // Don't use "fmt.Print*"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"regexp"
	"runtime"
	"time"

	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

// PluginType represents the plugin type
type PluginType int

// String returns string for matching enum plugin type
func (p PluginType) String() string {
	return types[p]
}

const (
	// CollectorPluginType - enum representation of collector plugin type
	CollectorPluginType PluginType = iota
	// ProcessorPluginType - enum representation of processor plugin type
	ProcessorPluginType
	// PublisherPluginType - enum representation of publisher plugin type
	PublisherPluginType
)

// PluginResponseState represents the plugin response states
type PluginResponseState int

const (
	// PluginSuccess - enum plugin response state of success
	PluginSuccess PluginResponseState = iota
	// PluginFailure - enum plugin response state of failure
	PluginFailure
)

// RPCType represents the enum type of RPC calls
type RPCType int

const (
	// NativeRPC - enum type of the native RPC
	NativeRPC RPCType = iota
	// JSONRPC - enum type of JSON RPC
	JSONRPC
)

var (
	// Timeout settings

	// PingTimeoutDurationDefault is how much time must elapse before a lack of Ping results in a timeout
	PingTimeoutDurationDefault = time.Millisecond * 1500
	// PingTimeoutLimit is how many succesive PingTimeouts must occur to equal a failure.
	PingTimeoutLimit = 3

	// Array matching plugin type enum to a string
	// note: in string represenation we use lower case
	types = [...]string{
		"collector",
		"processor",
		"publisher",
	}
)

// Plugin interface defines the types and methods
// that a plugin must implement
type Plugin interface {
	GetConfigPolicy() (*cpolicy.ConfigPolicy, error)
}

// PluginMeta for plugin
type PluginMeta struct {
	Name    string
	Version int
	Type    PluginType
	RPCType RPCType
	// Content types accepted by this plugin in priority order
	// snap.* means any snap type
	AcceptedContentTypes []string
	// Return content types in priority order
	// This is only really valid on processors
	ReturnedContentTypes []string
	// the max number of subscriptions this plugin
	// can handle
	ConcurrencyCount int
	// should always only be one instance of this plugin running
	Exclusive bool
	// do not encrypt communication with this plugin
	Unsecure bool
	// plugin cache TTL duration.
	// It will be converted from the client
	CacheTTL time.Duration
}

type metaOp func(m *PluginMeta)

// ConcurrencyCount sets the concurrent count in PluginMeta
// and returns the metaOp
func ConcurrencyCount(cc int) metaOp {
	return func(m *PluginMeta) {
		m.ConcurrencyCount = cc
	}
}

// Exclusive sets the exclusive flag in PluginMeta
func Exclusive(e bool) metaOp {
	return func(m *PluginMeta) {
		m.Exclusive = e
	}
}

// Unsecure sets Unsecure flag in the PluginMeta
func Unsecure(e bool) metaOp {
	return func(m *PluginMeta) {
		m.Unsecure = e
	}
}

// NewPluginMeta constructs and returns a PluginMeta struct
func NewPluginMeta(name string, version int, pluginType PluginType, acceptContentTypes, returnContentTypes []string, opts ...metaOp) *PluginMeta {
	// An empty accepted content type default to "snap.*"
	if len(acceptContentTypes) == 0 {
		acceptContentTypes = append(acceptContentTypes, "snap.*")
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

	p := &PluginMeta{
		Name:                 name,
		Version:              version,
		Type:                 pluginType,
		AcceptedContentTypes: acceptContentTypes,
		ReturnedContentTypes: returnContentTypes,

		//set the default for concurrency count to 1
		ConcurrencyCount: 1,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Arg struct type defines arguments passed to startup of Plugin
type Arg struct {
	// Plugin file path to binary
	PluginLogPath string
	// Ping timeout duration
	PingTimeoutDuration time.Duration

	NoDaemon bool
	// The listen port
	listenPort string
}

// NewArg returns a new instance of Arg passed to the startup of a plugin
func NewArg(logpath string) Arg {
	return Arg{
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
	PublicKey    *rsa.PublicKey
}

// Start starts a plugin where:
// PluginMeta - base information about plugin
// Plugin - either CollectorPlugin or PublisherPlugin
// requestString - plugins arguments (marshaled json of control/plugin Arg struct)
// returns an error and exitCode (exitCode from SessionState initilization or plugin termination code)
func Start(m *PluginMeta, c Plugin, requestString string) (error, int) {
	s, sErr, retCode := NewSessionState(requestString, c, m)
	if sErr != nil {
		return sErr, retCode
	}

	var (
		r        *Response
		exitCode int
	)

	switch m.Type {
	case CollectorPluginType:
		// Create our proxy
		proxy := &collectorPluginProxy{
			Plugin:  c.(CollectorPlugin),
			Session: s,
		}
		// Register the proxy under the "Collector" namespace
		rpc.RegisterName("Collector", proxy)

		r = &Response{
			Type:  CollectorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		if !m.Unsecure {
			r.PublicKey = &s.privateKey.PublicKey
		}
	case PublisherPluginType:
		r = &Response{
			Type:  PublisherPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		if !m.Unsecure {
			r.PublicKey = &s.privateKey.PublicKey
		}
		// Create our proxy
		proxy := &publisherPluginProxy{
			Plugin:  c.(PublisherPlugin),
			Session: s,
		}

		// Register the proxy under the "Publisher" namespace
		rpc.RegisterName("Publisher", proxy)
	case ProcessorPluginType:
		r = &Response{
			Type:  ProcessorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		if !m.Unsecure {
			r.PublicKey = &s.privateKey.PublicKey
		}
		// Create our proxy
		proxy := &processorPluginProxy{
			Plugin:  c.(ProcessorPlugin),
			Session: s,
		}
		// Register the proxy under the "Publisher" namespace
		rpc.RegisterName("Processor", proxy)
	}

	// Register common plugin methods used for utility reasons
	e := rpc.Register(s)
	if e != nil {
		if e.Error() != "rpc: service already defined: SessionState" {
			log.Println(e.Error())
			s.Logger().Println(e.Error())
			return e, 2
		}
	}

	l, err := net.Listen("tcp", "127.0.0.1:"+s.ListenPort())
	if err != nil {
		s.Logger().Println(err.Error())
		panic(err)
	}
	s.SetListenAddress(l.Addr().String())
	s.Logger().Printf("Listening %s\n", l.Addr())
	s.Logger().Printf("Session token %s\n", s.Token())

	switch r.Meta.RPCType {
	case JSONRPC:
		rpc.HandleHTTP()
		http.HandleFunc("/rpc", func(w http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			if req.ContentLength == 0 {
				encoder := json.NewEncoder(w)
				encoder.Encode(&struct {
					Id     interface{} `json:"id"`
					Result interface{} `json:"result"`
					Error  interface{} `json:"error"`
				}{
					Id:     nil,
					Result: nil,
					Error:  "rpc: method request ill-formed",
				})
				return
			}
			res := NewRPCRequest(req.Body).Call()
			io.Copy(w, res)
		})
		go http.Serve(l, nil)
	case NativeRPC:
		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					panic(err)
				}
				go rpc.ServeConn(conn)
			}
		}()
	default:
		panic("Unsupported RPC type")
	}

	resp := s.generateResponse(r)
	// Output response to stdout
	fmt.Println(string(resp))

	go s.heartbeatWatch(s.KillChan())

	if s.isDaemon() {
		exitCode = <-s.KillChan() // Closing of channel kills
	}

	return nil, exitCode
}

// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r    io.Reader     // holds the JSON formated RPC request
	rw   io.ReadWriter // holds the JSON formated RPC response
	done chan bool     // signals then end of the RPC request
}

// NewRPCRequest returns a new rpcRequest.
func NewRPCRequest(r io.Reader) *rpcRequest {
	var buf bytes.Buffer
	done := make(chan bool)
	return &rpcRequest{r, &buf, done}
}

// Read implements the io.ReadWriteCloser Read method.
func (r *rpcRequest) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

// Write implements the io.ReadWriteCloser Write method.
func (r *rpcRequest) Write(p []byte) (n int, err error) {
	n, err = r.rw.Write(p)
	defer func(done chan bool) { done <- true }(r.done)
	return
}

// Close implements the io.ReadWriteCloser Close method.
func (r *rpcRequest) Close() error {
	return nil
}

// Call invokes the RPC request, waits for it to complete, and returns the results.
func (r *rpcRequest) Call() io.Reader {
	go jsonrpc.ServeConn(r)
	<-r.done
	return r.rw
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
