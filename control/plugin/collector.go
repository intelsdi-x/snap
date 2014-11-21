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
	lf, err := os.OpenFile("/tmp/pulse_plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	defer lf.Close()
	pluginLog := log.New(lf, ">>>", log.Ldate|log.Ltime)
	// TODO

	pluginLog.Printf("\n")
	pluginLog.Printf("Starting collector plugin\n")
	if len(os.Args) < 2 {
		log.Fatalln("Pulse plugins are not started individually.")
		os.Exit(9)
	}

	var sessionState = new(SessionState)
	sessionState = InitSessionState(os.Args[0], os.Args[1])

	// Generate response
	// We should share as much as possible here.

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
		// Output response to stdout
		fmt.Print(string(resp))

		if err != nil {
			panic(err)
		}

		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					panic(err)
				}
				go rpc.ServeConn(conn)
			}
		}()

		// Right now we kill after 30 seconds until heartbeat is implemented
		time.Sleep(time.Second * 30)
	} else {
		sessionState.ListenAddress = ""
		r := Response{
			Type:  CollectorPluginType,
			State: PluginSuccess,
			Meta:  *m,
		}
		resp := sessionState.GenerateResponse(r)
		fmt.Print(string(resp))
		os.Exit(0)
	}
}
