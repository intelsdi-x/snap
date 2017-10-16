/*          **  DEPRECATED  **
For more information, see our deprecation notice
on Github: https://github.com/intelsdi-x/snap/issues/1289
*/

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
	"fmt"
	"net"
	"net/rpc"
	"regexp"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// How many successive PingTimeouts must occur to equal a failure.
	PingTimeoutLimit = 3
)

type metaOp func(m *PluginMeta)

// ConcurrencyCount is an option that can be be provided to the func NewPluginMeta.
func ConcurrencyCount(cc int) metaOp {
	return func(m *PluginMeta) {
		m.ConcurrencyCount = cc
	}
}

// Exclusive is an option that can be be provided to the func NewPluginMeta.
func Exclusive(e bool) metaOp {
	return func(m *PluginMeta) {
		m.Exclusive = e
	}
}

// Unsecure is an option that can be be provided to the func NewPluginMeta.
func Unsecure(e bool) metaOp {
	return func(m *PluginMeta) {
		m.Unsecure = e
	}
}

// RoutingStrategy is an option that can be be provided to the func NewPluginMeta.
func RoutingStrategy(r RoutingStrategyType) metaOp {
	return func(m *PluginMeta) {
		m.RoutingStrategy = r
	}
}

// CacheTTL is an option that can be be provided to the func NewPluginMeta.
func CacheTTL(t time.Duration) metaOp {
	return func(m *PluginMeta) {
		m.CacheTTL = t
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

// Start starts a plugin where:
// PluginMeta - base information about plugin
// Plugin - CollectorPlugin, ProcessorPlugin or PublisherPlugin
// requestString - plugins arguments (marshaled json of control/plugin Arg struct)
// returns an error and exitCode (exitCode from SessionState initialization or plugin termination code)
func Start(m *PluginMeta, c Plugin, requestString string) (error, int) {
	s, sErr, retCode := NewSessionState(requestString, c, m)
	if sErr != nil {
		return sErr, retCode
	}

	var (
		r        *Response
		exitCode int = 0
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
			s.Logger().Error(e.Error())
			return e, 2
		}
	}

	l, err := net.Listen("tcp", "127.0.0.1:"+s.ListenPort())
	if err != nil {
		s.Logger().Error(err.Error())
		panic(err)
	}
	s.SetListenAddress(l.Addr().String())
	s.Logger().Debugf("Listening %s\n", l.Addr())
	s.Logger().Debugf("Session token %s\n", s.Token())

	switch r.Meta.RPCType {
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
	s.Logger().Println(string(resp))
	go s.heartbeatWatch(s.KillChan())

	if s.isDaemon() {
		exitCode = <-s.KillChan() // Closing of channel kills
	}

	return nil, exitCode
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
