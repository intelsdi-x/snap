package rest

import (
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
	log.WithFields(log.Fields{
		"module": "mgmt-rest",
		"index":  l.counter,
		"method": r.Method,
		"url":    r.URL.Path,
	}).Info("API request")
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
