package plugin

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core/ctypes"
)

// Publisher plugin
type PublisherPlugin interface {
	Plugin
	Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue, logger *log.Logger) error
	GetConfigPolicyNode() cpolicy.ConfigPolicyNode
}

func StartPublisher(p PublisherPlugin, s Session, r *Response) (error, int) {
	var exitCode int = 0

	l, err := net.Listen("tcp", "127.0.0.1:"+s.ListenPort())
	if err != nil {
		s.Logger().Println(err.Error())
		panic(err)
	}
	s.SetListenAddress(l.Addr().String())
	s.Logger().Printf("Listening %s\n", l.Addr())
	s.Logger().Printf("Session token %s\n", s.Token())

	// Create our proxy
	proxy := &publisherPluginProxy{
		Plugin:  p,
		Session: s,
	}

	// Register the proxy under the "Publisher" namespace
	rpc.RegisterName("Publisher", proxy)
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

func init() {
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
}
