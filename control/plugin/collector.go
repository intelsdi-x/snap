package plugin

import (
	"fmt"
	"log" // TODO proper logging to file or elsewhere
	"net"
	"net/rpc"
	"os"
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
		resp := sessionState.GenerateResponse(l.Addr().String())
		// time.Sleep(time.Second * 2)
		fmt.Print(string(resp))

		if err != nil {
			panic(err)
		}

		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}
			go rpc.ServeConn(conn)
		}
	} else {
		resp := sessionState.GenerateResponse("")
		fmt.Print(string(resp))
		os.Exit(0)
	}
}
