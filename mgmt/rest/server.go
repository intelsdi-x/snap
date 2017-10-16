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
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"strings"

	"github.com/intelsdi-x/snap/mgmt/rest/api"
	"github.com/intelsdi-x/snap/mgmt/rest/v1"
	"github.com/intelsdi-x/snap/mgmt/rest/v2"
)

const (
	allowedMethods = "GET, POST, DELETE, PUT, OPTIONS"
	allowedHeaders = "Origin, X-Requested-With, Content-Type, Accept"
	maxAge         = 3600
)

var (
	ErrBadCert = errors.New("Invalid certificate given")

	restLogger     = log.WithField("_module", "_mgmt-rest")
	protocolPrefix = "http"
)

type Server struct {
	apis           []api.API
	n              *negroni.Negroni
	r              *httprouter.Router
	snapTLS        *snapTLS
	auth           bool
	pprof          bool
	authpwd        string
	addrString     string
	addr           net.Addr
	wg             sync.WaitGroup
	killChan       chan struct{}
	err            chan error
	allowedOrigins map[string]bool
	// the following instance variables are used to cleanly shutdown the server
	serverListener net.Listener
	closingChan    chan bool
}

// New creates a REST API server with a given config
func New(cfg *Config) (*Server, error) {
	// pull a few parameters from the configuration passed in by snapteld
	s := &Server{
		err:        make(chan error),
		killChan:   make(chan struct{}),
		addrString: cfg.Address,
		pprof:      cfg.Pprof,
	}
	if cfg.HTTPS {
		var err error
		s.snapTLS, err = newtls(cfg.RestCertificate, cfg.RestKey)
		if err != nil {
			return nil, err
		}
		protocolPrefix = "https"
	}
	restLogger.Info(fmt.Sprintf("Configuring REST API with HTTPS set to: %v", cfg.HTTPS))

	s.apis = []api.API{
		v1.New(&s.wg, s.killChan, protocolPrefix),
		v2.New(&s.wg, s.killChan, protocolPrefix),
	}

	s.n = negroni.New(
		NewLogger(),
		negroni.NewRecovery(),
		negroni.HandlerFunc(s.authMiddleware),
	)
	s.r = httprouter.New()

	// CORS has to be turned on explicitly in the global config.
	// Otherwise, it defauts to the same origin.
	origins, err := s.getAllowedOrigins(cfg.Corsd)
	if err != nil {
		return nil, err
	}
	if len(origins) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins: origins,
			AllowedMethods: []string{allowedMethods},
			AllowedHeaders: []string{allowedHeaders},
			MaxAge:         maxAge,
		})
		s.n.Use(c)
	}

	// Use negroni to handle routes
	s.n.UseHandler(s.r)
	return s, nil
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
	reqOrigin := r.Header.Get("Origin")
	s.setAllowedOrigins(rw, reqOrigin)

	defer r.Body.Close()
	if s.auth {
		_, password, ok := r.BasicAuth()
		// If we have valid password or going to tribe/agreements endpoint
		// go to next. tribe/agreements endpoint used for populating
		// snaptel help page when tribe mode is turned on.
		if ok && password == s.authpwd {
			next(rw, r)
		} else {
			v2.Write(401, v2.UnauthError{Code: 401, Message: "Not authorized. Please specify the same password that used to start snapteld. E.g: [snaptel -p plugin list] or [curl http://localhost:8181/v2/plugins -u snap]"}, rw)
		}
	} else {
		next(rw, r)
	}
}

// CORS origins have to be turned on explicitly in the global config.
// Otherwise, it defaults to the same origin.
func (s *Server) setAllowedOrigins(rw http.ResponseWriter, ro string) {
	if len(s.allowedOrigins) > 0 {
		if _, ok := s.allowedOrigins[ro]; ok {
			// localhost CORS is not supported by all browsers. It has to use "*".
			if strings.Contains(ro, "127.0.0.1") || strings.Contains(ro, "localhost") {
				ro = "*"
			}
			rw.Header().Set("Access-Control-Allow-Origin", ro)
			rw.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			rw.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			rw.Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
		}
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
			log.Fatal(err)
		}
		s.serverListener = ln
		s.wg.Add(1)
		go s.serveTLS(ln)
	} else {
		ln, err := net.Listen("tcp", addrString)
		if err != nil {
			log.Fatal(err)
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

func (s *Server) addRoutes() {
	for _, apiInstance := range s.apis {
		for _, route := range apiInstance.GetRoutes() {
			s.r.Handle(route.Method, route.Path, route.Handle)
		}
	}
	s.addPprofRoutes()
}

func (s *Server) getAllowedOrigins(corsd string) ([]string, error) {
	// Avoids panics when validating URLs.
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
				fmt.Println(err)
			}
		}

	}()

	if corsd == "" {
		return []string{}, nil
	}

	vo := []string{}
	s.allowedOrigins = map[string]bool{}

	os := strings.Split(corsd, ",")
	for _, o := range os {
		to := strings.TrimSpace(o)

		// Validates origin formation
		u, err := url.Parse(to)

		// Checks if scheme or host exists when no error occurred.
		if err != nil || u.Scheme == "" || u.Host == "" {
			restLogger.Errorf("Invalid origin found %s", to)
			return []string{}, fmt.Errorf("Invalid origin found: %s.", to)
		}

		vo = append(vo, to)
		s.allowedOrigins[to] = true
	}
	return vo, nil
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
