package plugin

import (
	"fmt"
	"log" // TODO proper logging to file or elsewhere
	"net"
	"net/rpc"
	"os"
	"time"
)

// Collector plugin
type CollectorPlugin interface {
	Plugin
	Collect(CollectorArgs, *CollectorReply) error
}

// Arguments passed to Collect() for a Collector implementation
type CollectorArgs struct {
}

// Reply assigned by a Collector implementation using Collect()
type CollectorReply struct {
}

// Execution method for a Collector plugin.
func StartCollector(m *PluginMeta, c CollectorPlugin, p *ConfigPolicy) {

	// TODO - Patching in logging, needs to be replaced with proper log pathing and deterministic log file naming

	// defer lf.Close()
	// pluginLog := log.New(os.Stdout, ">>>", log.Ldate|log.Ltime)
	// pluginLog := log.New(lf, ">>>", log.Ldate|log.Ltime)
	// TODO
	//
	// pluginLog.Printf("\n")
	// pluginLog.Printf("Starting collector plugin\n")
	// if len(os.Args) < 2 {
	// log.Fatalln("Pulse plugins are not started individually.")
	// os.Exit(9)
	// }
	//
	// pluginLog.Println(os.Args[0])
	// pluginLog.Println(os.Args[1])
	sessionState, sErr := InitSessionState(os.Args[0], os.Args[1])
	if sErr != nil {
		// fmt.Printf("error parsing arguments: %s\n", sErr.Error())
		os.Exit(1)
	}
	var pluginLog *log.Logger
	switch lp := sessionState.Arg.PluginLogPath; lp {
	case "", "/tmp":
		// Empty means use default tmp log (needs to be removed post-alpha)
		lf, err := os.OpenFile("/tmp/pulse_plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// fmt.Printf("error opening log file: %v", err)
			os.Exit(1)
		}
		pluginLog = log.New(lf, ">>>", log.Ldate|log.Ltime)
	default:
		lf, err := os.OpenFile(lp, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// fmt.Printf("error opening log file (%s): %v", lp, err)
			os.Exit(1)
		}
		pluginLog = log.New(lf, ">>>", log.Ldate|log.Ltime)
	}

	// if sErr != nil {
	// pluginLog.Println(sErr)
	// }
	sessionState.LastPing = time.Now()

	// Generate response
	// We should share as much as possible here.

	pluginLog.Printf("Daemon mode: %t\n", sessionState.RunAsDaemon)
	// if not in daemon mode we don't need to setup listener
	if sessionState.RunAsDaemon {
		// Register the collector RPC methods from plugin implementation
		rpc.RegisterName("collector", c)
		// Register common plugin methods used for utility reasons
		rpc.RegisterName("meta", m)

		// Right now we only listen on TCP connections. Optionally consider a UNIX socket option.
		l, err := net.Listen("tcp", "127.0.0.1:"+sessionState.ListenPort)
		pluginLog.Printf("Listening %s\n", l.Addr())
		pluginLog.Printf("Session token %s\n", sessionState.Token)
		// Add the listening information to the session state
		sessionState.ListenAddress = l.Addr().String()

		// Generate a response
		r := Response{
			Type:  CollectorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		resp := sessionState.GenerateResponse(r)
		pluginLog.Println(string(resp))
		// Output response to stdout
		fmt.Println(string(resp))

		// Start ping listener
		// If it has not received a ping in N amount of time * T it quits.
		killChan := make(chan interface{})

		pluginLog.Println("Watching Ping timeout")
		go watchLastPing(killChan, sessionState, pluginLog)

		if err != nil {
			pluginLog.Println(err.Error())
			panic(err)
		}

		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					panic(err)
				}

				defer conn.Close()
				go rpc.ServeConn(conn)
			}
		}()

		<-killChan // Closing of channel kills
	} else {
		sessionState.ListenAddress = ""
		r := Response{
			Type:  CollectorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		resp := sessionState.GenerateResponse(r)
		fmt.Print(string(resp))
	}
	pluginLog.Println("Exiting!")
	os.Exit(0)
}

func watchLastPing(killChan chan (interface{}), s *SessionState, l *log.Logger) {
	l.Println("Watching Ping timeout")
	count := 0
	for {
		if time.Now().Sub(s.LastPing) >= PingTimeoutDuration {
			l.Println("Ping timeout fired")
			count++
			if count >= PingTimeoutLimit {
				l.Println("Ping timeout expired")
				defer close(killChan)
				return
			}
		} else {
			// Reset count
			count = 0
		}
		time.Sleep(PingTimeoutDuration)
		l.Println("Ping timeout tick")
	}
}
