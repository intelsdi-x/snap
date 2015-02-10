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
	GetMetricTypes(GetMetricTypesArgs, *GetMetricTypesReply) error
}

// Arguments passed to Collect() for a Collector implementation
type CollectorArgs struct {
}

// Reply assigned by a Collector implementation using Collect()
type CollectorReply struct {
}

type GetMetricTypesArgs struct {
}

type GetMetricTypesReply struct {
	MetricTypes []*MetricType
}

// Execution method for a Collector plugin.
func StartCollector(m *PluginMeta, c CollectorPlugin, p *ConfigPolicy) error {

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
		return sErr
	}

	switch lp := sessionState.Arg.PluginLogPath; lp {
	case "", "/tmp":
		// Empty means use default tmp log (needs to be removed post-alpha)
		lf, err := os.OpenFile("/tmp/pulse_plugin.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// fmt.Printf("error opening log file: %v", err)
			os.Exit(1)
		}
		sessionState.Logger = log.New(lf, ">>>", log.Ldate|log.Ltime)
	default:
		lf, err := os.OpenFile(lp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			// fmt.Printf("error opening log file (%s): %v", lp, err)
			os.Exit(1)
		}
		sessionState.Logger = log.New(lf, ">>>", log.Ldate|log.Ltime)
	}
	sessionState.LastPing = time.Now()

	// We register RPC even in non-daemon mode to ensure it would be successful.
	// Register the collector RPC methods from plugin implementation
	rpc.Register(c)
	// Register common plugin methods used for utility reasons
	e := rpc.Register(sessionState)
	// If the rpc registration has an error we need to halt.
	if e != nil {
		log.Println(e.Error())
		sessionState.Logger.Println(e.Error())
		return e
	}
	// Generate response
	r := Response{
		Type:  CollectorPluginType,
		State: PluginSuccess,
		Meta:  *m,
	}
	sessionState.Logger.Printf("Daemon mode: %t\n", sessionState.RunAsDaemon)
	// if not in daemon mode we don't need to setup listener
	if sessionState.RunAsDaemon {
		// Right now we only listen on TCP connections. Optionally consider a UNIX socket option.
		l, err := net.Listen("tcp", "127.0.0.1:"+sessionState.ListenPort)
		sessionState.Logger.Printf("Listening %s\n", l.Addr())
		sessionState.Logger.Printf("Session token %s\n", sessionState.Token)
		// Add the listening information to the session state
		sessionState.ListenAddress = l.Addr().String()

		// Generate a response
		resp := generateResponse(r, sessionState)
		sessionState.Logger.Println(string(resp))
		// Output response to stdout
		fmt.Println(string(resp))

		// Start ping listener
		// If it has not received a ping in N amount of time * T it quits.
		killChan := make(chan struct{})
		go sessionState.heartbeatWatch(killChan)

		if err != nil {
			sessionState.Logger.Println(err.Error())
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

		<-killChan // Closing of channel kills
	} else {
		sessionState.ListenAddress = ""
		resp := generateResponse(r, sessionState)
		fmt.Print(string(resp))
	}
	sessionState.Logger.Println("Exiting!")
	return nil
}
