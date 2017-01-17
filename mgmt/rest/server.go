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

package rest

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"

	"github.com/intelsdi-x/snap/mgmt/rest/v1"
	"github.com/intelsdi-x/snap/mgmt/rest/v2"
	"github.com/intelsdi-x/snap/pkg/api"
)

// default configuration values
const (
	defaultEnable          bool   = true
	defaultPort            int    = 8181
	defaultAddress         string = ""
	defaultHTTPS           bool   = false
	defaultRestCertificate string = ""
	defaultRestKey         string = ""
	defaultAuth            bool   = false
	defaultAuthPassword    string = ""
	defaultPortSetByConfig bool   = false
	defaultPprof           bool   = false
)

var (
	ErrBadCert = errors.New("Invalid certificate given")

	restLogger     = log.WithField("_module", "_mgmt-rest")
	protocolPrefix = "http"
)

// holds the configuration passed in through the SNAP config file
//   Note: if this struct is modified, then the switch statement in the
//         UnmarshalJSON method in this same file needs to be modified to
//         match the field mapping that is defined here
type Config struct {
	Enable           bool   `json:"enable"yaml:"enable"`
	Port             int    `json:"port"yaml:"port"`
	Address          string `json:"addr"yaml:"addr"`
	HTTPS            bool   `json:"https"yaml:"https"`
	RestCertificate  string `json:"rest_certificate"yaml:"rest_certificate"`
	RestKey          string `json:"rest_key"yaml:"rest_key"`
	RestAuth         bool   `json:"rest_auth"yaml:"rest_auth"`
	RestAuthPassword string `json:"rest_auth_password"yaml:"rest_auth_password"`
	portSetByConfig  bool   ``
	Pprof            bool   `json:"pprof"yaml:"pprof"`
}

const (
	CONFIG_CONSTRAINTS = `
			"restapi" : {
				"type": ["object", "null"],
				"properties" : {
					"enable": {
						"type": "boolean"
					},
					"https" : {
						"type": "boolean"
					},
					"rest_auth": {
						"type": "boolean"
					},
					"rest_auth_password": {
						"type": "string"
					},
					"rest_certificate": {
						"type": "string"
					},
					"rest_key" : {
						"type": "string"
					},
					"port" : {
						"type": "integer",
						"minimum": 1,
						"maximum": 65535
					},
					"addr" : {
						"type": "string"
					},
					"pprof": {
						"type": "boolean"
					}
				},
				"additionalProperties": false
			}
	`
)

type Server struct {
	apis       []api.API
	n          *negroni.Negroni
	r          *httprouter.Router
	snapTLS    *snapTLS
	auth       bool
	pprof      bool
	authpwd    string
	addrString string
	addr       net.Addr
	wg         sync.WaitGroup
	killChan   chan struct{}
	err        chan error
	// the following instance variables are used to cleanly shutdown the server
	serverListener net.Listener
	closingChan    chan bool
}

// New creates a REST API server with a given config
func New(cfg *Config) (*Server, error) {
	// pull a few parameters from the configuration passed in by snapteld
	https := cfg.HTTPS
	cpath := cfg.RestCertificate
	kpath := cfg.RestKey
	pprof := cfg.Pprof
	s := &Server{
		err:        make(chan error),
		killChan:   make(chan struct{}),
		addrString: cfg.Address,
		pprof:      pprof,
	}
	if https {
		var err error
		s.snapTLS, err = newtls(cpath, kpath)
		if err != nil {
			return nil, err
		}
		protocolPrefix = "https"
	}

	s.apis = []api.API{
		v1.NewV1(&s.wg, s.killChan, protocolPrefix),
		v2.NewV2(&s.wg, s.killChan, protocolPrefix),
	}

	restLogger.Info(fmt.Sprintf("Configuring REST API with HTTPS set to: %v", https))
	s.n = negroni.New(
		NewLogger(),
		negroni.NewRecovery(),
		negroni.HandlerFunc(s.authMiddleware),
	)
	s.r = httprouter.New()
	// Use negroni to handle routes
	s.n.UseHandler(s.r)
	return s, nil
}

// GetDefaultConfig gets the default snapteld configuration
func GetDefaultConfig() *Config {
	return &Config{
		Enable:           defaultEnable,
		Port:             defaultPort,
		Address:          defaultAddress,
		HTTPS:            defaultHTTPS,
		RestCertificate:  defaultRestCertificate,
		RestKey:          defaultRestKey,
		RestAuth:         defaultAuth,
		RestAuthPassword: defaultAuthPassword,
		portSetByConfig:  defaultPortSetByConfig,
		Pprof:            defaultPprof,
	}
}

// define a method that can be used to determine if the port the RESTful
// API is listening on was set in the configuration file
func (c *Config) PortSetByConfigFile() bool {
	return c.portSetByConfig
}

// UnmarshalJSON unmarshals valid json into a Config.  An example Config can be found
// at github.com/intelsdi-x/snap/blob/master/examples/configs/snap-config-sample.json
func (c *Config) UnmarshalJSON(data []byte) error {
	// construct a map of strings to json.RawMessages (to defer the parsing of individual
	// fields from the unmarshalled interface until later) and unmarshal the input
	// byte array into that map
	t := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	// loop through the individual map elements, parse each in turn, and set
	// the appropriate field in this configuration
	for k, v := range t {
		switch k {
		case "enable":
			if err := json.Unmarshal(v, &(c.Enable)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::enable')", err)
			}
		case "port":
			if err := json.Unmarshal(v, &(c.Port)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::port')", err)
			}
			c.portSetByConfig = true
		case "addr":
			if err := json.Unmarshal(v, &(c.Address)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::addr')", err)
			}
		case "https":
			if err := json.Unmarshal(v, &(c.HTTPS)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::https')", err)
			}
		case "rest_certificate":
			if err := json.Unmarshal(v, &(c.RestCertificate)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_certificate')", err)
			}
		case "rest_key":
			if err := json.Unmarshal(v, &(c.RestKey)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_key')", err)
			}
		case "rest_auth":
			if err := json.Unmarshal(v, &(c.RestAuth)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_auth')", err)
			}
		case "rest_auth_password":
			if err := json.Unmarshal(v, &(c.RestAuthPassword)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::rest_auth_password')", err)
			}
		case "pprof":
			if err := json.Unmarshal(v, &(c.Pprof)); err != nil {
				return fmt.Errorf("%v (while parsing 'restapi::pprof')", err)
			}
		default:
			return fmt.Errorf("Unrecognized key '%v' in global config file while parsing 'restapi'", k)
		}
	}
	return nil
}

// SetAPIAuth sets API authentication to enabled or disabled
func (s *Server) SetAPIAuth(auth bool) {
	s.auth = auth
}

// SetAPIAuthPwd sets the API authentication password from snapteld
func (s *Server) SetAPIAuthPwd(pwd string) {
	s.authpwd = pwd
}

// Auth Middleware for REST API
func (s *Server) authMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer r.Body.Close()
	if s.auth {
		_, password, ok := r.BasicAuth()
		// If we have valid password or going to tribe/agreements endpoint
		// go to next. tribe/agreements endpoint used for populating
		// snaptel help page when tribe mode is turned on.
		if ok && password == s.authpwd {
			next(rw, r)
		} else {
			http.Error(rw, "Not Authorized", 401)
		}
	} else {
		next(rw, r)
	}
}

func (s *Server) SetAddress(addrString string) {
	s.addrString = addrString
	restLogger.Info(fmt.Sprintf("Address used for binding: [%v]", s.addrString))
}

func (s *Server) Name() string {
	return "REST"
}

func (s *Server) Start() error {
	s.closingChan = make(chan bool, 1)
	s.addRoutes()
	s.run(s.addrString)
	restLogger.WithFields(log.Fields{
		"_block": "start",
	}).Info("REST started")
	return nil
}

func (s *Server) Stop() {
	// add a boolean to the s.closingChan (used for error handling in the
	// goroutine that is listening for connections)
	s.closingChan <- true
	// then close the server
	close(s.killChan)
	// close the server listener
	s.serverListener.Close()
	// wait for the server goroutines to complete (serve and watch)
	s.wg.Wait()
	// finally log the result
	restLogger.WithFields(log.Fields{
		"_block": "stop",
	}).Info("REST stopped")
}

func (s *Server) Err() <-chan error {
	return s.err
}

func (s *Server) Port() int {
	return s.addr.(*net.TCPAddr).Port
}

func (s *Server) run(addrString string) {
	restLogger.Info("Starting REST API on ", addrString)
	if s.snapTLS != nil {
		cer, err := tls.LoadX509KeyPair(s.snapTLS.cert, s.snapTLS.key)
		if err != nil {
			s.err <- err
			return
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err := tls.Listen("tcp", addrString, config)
		if err != nil {
			s.err <- err
		}
		s.serverListener = ln
		s.wg.Add(1)
		go s.serveTLS(ln)
	} else {
		ln, err := net.Listen("tcp", addrString)
		if err != nil {
			s.err <- err
		}
		s.serverListener = ln
		s.addr = ln.Addr()
		s.wg.Add(1)
		go s.serve(ln)
	}
}

func (s *Server) serveTLS(ln net.Listener) {
	defer s.wg.Done()
	err := http.Serve(ln, s.n)
	if err != nil {
		select {
		case <-s.closingChan:
		// If we called Stop() then there will be a value in s.closingChan, so
		// we'll get here and we can exit without showing the error.
		default:
			restLogger.Error(err)
			s.err <- err
		}
	}
}

func (s *Server) serve(ln net.Listener) {
	defer s.wg.Done()
	err := http.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}, s.n)
	if err != nil {
		select {
		case <-s.closingChan:
		// If we called Stop() then there will be a value in s.closingChan, so
		// we'll get here and we can exit without showing the error.
		default:
			restLogger.Error(err)
			s.err <- err
		}
	}
}

// Monkey patch ListenAndServe and TCP alive code from https://golang.org/src/net/http/server.go
// The built in ListenAndServe and ListenAndServeTLS include TCP keepalive
// At this point the Go team is not wanting to provide separate listen and serve methods
// that also provide an exported TCP keepalive per: https://github.com/golang/go/issues/12731
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func (s *Server) BindMetricManager(m api.Metrics) {
	for _, apiInstance := range s.apis {
		apiInstance.BindMetricManager(m)
	}
}

func (s *Server) BindTaskManager(t api.Tasks) {
	for _, apiInstance := range s.apis {
		apiInstance.BindTaskManager(t)
	}
}

func (s *Server) BindTribeManager(t api.Tribe) {
	for _, apiInstance := range s.apis {
		apiInstance.BindTribeManager(t)
	}
}

func (s *Server) BindConfigManager(c api.Config) {
	for _, apiInstance := range s.apis {
		apiInstance.BindConfigManager(c)
	}
}

func (s *Server) addRoutes() {
	for _, apiInstance := range s.apis {
		for _, route := range apiInstance.GetRoutes() {
			s.r.Handle(route.Method, route.Path, route.Handle)
		}
	}
	// profiling tools routes
	if s.pprof {
		s.r.GET("/debug/pprof/", s.index)
		s.r.GET("/debug/pprof/block", s.index)
		s.r.GET("/debug/pprof/goroutine", s.index)
		s.r.GET("/debug/pprof/heap", s.index)
		s.r.GET("/debug/pprof/threadcreate", s.index)
		s.r.GET("/debug/pprof/cmdline", s.cmdline)
		s.r.GET("/debug/pprof/profile", s.profile)
		s.r.GET("/debug/pprof/symbol", s.symbol)
		s.r.GET("/debug/pprof/trace", s.trace)
	}
}

// profiling tools handlers

func (s *Server) index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Index(w, r)
}

func (s *Server) cmdline(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Cmdline(w, r)
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Profile(w, r)
}

func (s *Server) symbol(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Symbol(w, r)
}

func (s *Server) trace(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pprof.Trace(w, r)
}
