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

type CollectorArgs struct {
}

type CollectorReply struct {
}

// A collector plugin and a config policy for it
func StartCollector(m *PluginMeta, c CollectorPlugin, p *ConfigPolicy) {

	lf, err := os.OpenFile("/tmp/pulse_plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	defer lf.Close()
	pluginLog := log.New(lf, ">>>", log.Ldate|log.Ltime)

	// Init plugin session state
	// pluginState := InitSessionState()

	//
	pluginLog.Printf("Starting collector plugin\n")
	if len(os.Args) < 2 {
		log.Fatalln("Pulse plugins are not started individually.")
		os.Exit(9)
	}

	var sessionState = new(SessionState)
	sessionState = InitSessionState(os.Args[0], os.Args[1])

	rpc.RegisterName("collector", c)
	rpc.RegisterName("meta", m)

	l, err := net.Listen("tcp", "127.0.0.1:"+sessionState.ListenPort)
	pluginLog.Printf("Listening %s\n", l.Addr())
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
}
