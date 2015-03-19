package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type LogLevel int

var (
	mutex  = sync.Mutex{}
	level  = DebugLevel
	Output io.Writer
	logger *log.Logger
)

const (
	DebugLevel LogLevel = iota + 1
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

func lazyLoad() {
	if logger != nil {
		return
	}
	if Output != nil {
		logger = log.New(Output, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}
}

func SetLevel(l LogLevel) {
	mutex.Lock()
	defer mutex.Unlock()
	level = l
}

func Debug(s ...string) {
	lazyLoad()
	if level == DebugLevel {
		logger.Printf("DEBUG: %s", strings.Join(s, " - "))
	}
}

func Debugf(p string, b string, a ...interface{}) {
	lazyLoad()
	if level <= DebugLevel {
		logger.Printf(fmt.Sprintf("DEBUG: %s - %s", p, b), a...)
	}
}

func Info(s ...string) {
	lazyLoad()
	if level <= InfoLevel {
		logger.Printf("INFO: %s", strings.Join(s, " - "))
	}
}

func Infof(p string, b string, a ...interface{}) {
	lazyLoad()
	if level <= InfoLevel {
		logger.Printf(fmt.Sprintf("INFO: %s - %s", p, b), a...)
	}
}

func Warning(s ...string) {
	lazyLoad()
	if level <= WarningLevel {
		logger.Printf("WARNING: %s", strings.Join(s, " - "))
	}
}

func Warningf(p string, b string, a ...interface{}) {
	lazyLoad()
	if level <= WarningLevel {
		logger.Printf(fmt.Sprintf("Warning: %s - %s", p, b), a...)
	}
}

func Error(s ...string) {
	lazyLoad()
	if level <= ErrorLevel {
		logger.Printf("ERROR: %s", strings.Join(s, " - "))
	}
}

func Errorf(p string, b string, a ...interface{}) {
	lazyLoad()
	if level <= ErrorLevel {
		logger.Printf(fmt.Sprintf("Error: %s - %s", p, b), a...)
	}
}

func Fatal(s ...string) {
	lazyLoad()
	if level <= FatalLevel {
		logger.Printf("FATAL: %s", strings.Join(s, " - "))
	}
}

func Fatalf(p string, b string, a ...interface{}) {
	lazyLoad()
	if level <= FatalLevel {
		logger.Printf(fmt.Sprintf("Fatal: %s - %s", p, b), a...)
	}
}
