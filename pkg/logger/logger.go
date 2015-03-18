package logger

import (
	// "fmt"
	"io"
	"log"
	"strings"
	"sync"
)

type LogLevel int

var (
	mutex = sync.Mutex{}
	level = DebugLevel
)

const (
	DebugLevel LogLevel = iota + 1
	InfoLevel
	WarningLevel
	ErrorLevel
	FatalLevel
)

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func SetLevel(l LogLevel) {
	mutex.Lock()
	defer mutex.Unlock()
	level = l
}

func Debug(s ...string) {
	if level == DebugLevel {
		log.Printf("DEBUG: %s", strings.Join(s, " - "))
	}
}

func Info(s ...string) {
	if level <= InfoLevel {
		log.Printf("INFO: %s", strings.Join(s, " - "))
	}
}

func Warning(s ...string) {
	if level <= WarningLevel {
		log.Printf("WARNING: %s", strings.Join(s, " - "))
	}
}

func Error(s ...string) {
	if level <= ErrorLevel {
		log.Printf("ERROR: %s", strings.Join(s, " - "))
	}
}

func Fatal(s ...string) {
	if level <= FatalLevel {
		log.Printf("FATAL: %s", strings.Join(s, " - "))
	}
}
