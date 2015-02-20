package plugin

import (
	"fmt"
	"log" // TODO proper logging to file or elsewhere
	"net"
	"net/rpc"
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

// GetMetricTypesArgs args passed to GetMetricTypes
type GetMetricTypesArgs struct {
}

// GetMetricTypesReply assigned by GetMetricTypes() implementation
type GetMetricTypesReply struct {
	MetricTypes []*MetricType
}

// Execution method for a Collector plugin. Error and exit code (int) returned.
func StartCollector(c CollectorPlugin, s Session, r *Response) (error, int) {
	var exitCode int = 0

	l, err := net.Listen("tcp", "127.0.0.1:"+s.ListenPort())
	if err != nil {
		s.Logger().Println(err.Error())
		panic(err)
	}
	s.SetListenAddress(l.Addr().String())
	s.Logger().Printf("Listening %s\n", l.Addr())
	s.Logger().Printf("Session token %s\n", s.Token())

	// We register RPC even in non-daemon mode to ensure it would be successful.
	// Register the collector RPC methods from plugin implementation
	rpc.Register(c)
	// Register common plugin methods used for utility reasons
	e := rpc.Register(s)
	if e != nil {
		if e.Error() != "rpc: service already defined: SessionState" {
			log.Println(e.Error())
			s.Logger().Println(e.Error())
			return e, 2
		}
	}

	resp := s.generateResponse(r)
	// Output response to stdout
	fmt.Println(string(resp))

	go s.heartbeatWatch(s.KillChan())

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				panic(err)
			}
			go rpc.ServeConn(conn)
		}
	}()

	if s.isDaemon() {
		exitCode = <-s.KillChan() // Closing of channel kills
	}

	return nil, exitCode
}
