package plugin

import (
	"log" // TODO proper logging to file or elsewhere

	"net"
	"net/rpc"
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
func StartCollector(n string, v int, c CollectorPlugin, p *ConfigPolicy) {
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
