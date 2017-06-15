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
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
)

// Logger is a snap middleware that logs to a logrus facility
type Logger struct {
	counter uint
}

// NewLogger returns a new Logger instance
func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.counter++
	restLogger.WithFields(log.Fields{
		"index":  l.counter,
		"method": r.Method,
		"url":    r.URL.Path,
	}).Debug("API request")
	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	restLogger.WithFields(log.Fields{
		"index":       l.counter,
		"method":      r.Method,
		"url":         r.URL.Path,
		"status-code": res.Status(),
		"status":      http.StatusText(res.Status()),
	}).Debug("API response")

	if deprecationInfo := rw.Header().Get("Deprecated"); len(deprecationInfo) != 0 {
		restLogger.WithFields(log.Fields{
			"method": r.Method,
			"url":    r.URL.Path,
		}).Warning(deprecationInfo)
	}

}
