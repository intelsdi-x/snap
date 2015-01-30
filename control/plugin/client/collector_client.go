package client

// "net"
// "net/rpc"
// "time"

// func NewCollectorClient(address string, timeout time.Duration) (*PluginNativeClient, error) {
// 	// Attempt to dial address error on timeout or problem
// 	conn, err := net.DialTimeout("tcp", address, timeout)
// 	// Return nil RPCClient and err if encoutered
// 	if err != nil {
// 		return nil, err
// 	}
// 	r := rpc.NewClient(conn)
// 	p := &PluginNativeClient{connection: r}
// 	return p, nil
// }
