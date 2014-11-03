package plugin

// Config Policy
// task > control > default

import (
	"log" // TODO proper logging to file or elsewhere

	"net"
	"net/rpc"
	// pulserpc "github.com/intelsdilabs/pulse/control/plugin/rpc"
)

// Plugin interface
type Plugin interface {
}

// Collector plugin
type CollectorPlugin interface {
	Plugin
}

// Publisher plugin
type PublisherPlugin interface {
	Plugin
}

type ConfigPolicy struct {
}

func NewServer() (*Server, error) {
	return new(Server), nil
}

type Server struct {
}

// A collector plugin and a config policy for it
func (s *Server) StartCollector(n string, v int, c CollectorPlugin, p *ConfigPolicy) {
	//
	log.Printf("Starting collector plugin\n")

	rpc.RegisterName("collector", c)
	l, err := net.Listen("tcp", "127.0.0.1:30001")
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
