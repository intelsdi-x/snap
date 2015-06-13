package rest

import (
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

// Logger is a Pulse middleware that logs to a logrus facility
type Logger struct {
	counter uint
}

// NewLogger returns a new Logger instance
func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.counter++
	// For debug level we output more fields and request details including body
	if log.GetLevel() == log.DebugLevel {
		b, _ := ioutil.ReadAll(r.Body)
		f := log.Fields{
			"module": "mgmt-rest",
			"index":  l.counter,
			"method": r.Method,
			"url":    r.URL.Path,
			"body":   string(b),
		}
		for k, v := range r.Header {
			f[k] = v
		}
		log.WithFields(f).Debug("API request")
	} else {
		log.WithFields(log.Fields{
			"module": "mgmt-rest",
			"index":  l.counter,
			"method": r.Method,
			"url":    r.URL.Path,
		}).Info("API request")
	}

	next(rw, r)
	res := rw.(negroni.ResponseWriter)
	log.WithFields(log.Fields{
		"module":      "mgmt-rest",
		"index":       l.counter,
		"method":      r.Method,
		"url":         r.URL.Path,
		"status-code": res.Status(),
		"status":      http.StatusText(res.Status()),
	}).Info("API response")
}
