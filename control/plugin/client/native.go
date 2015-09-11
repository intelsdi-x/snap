package client

import (
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"net"
	"net/rpc"
	"time"
	"unicode"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/control/plugin/encoding"
	"github.com/intelsdi-x/pulse/control/plugin/encrypter"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/ctypes"
)

// CallsRPC provides an interface for RPC clients
type CallsRPC interface {
	Call(methd string, args interface{}, reply interface{}) error
}

// Native clients use golang net/rpc for communication to a native rpc server.
type PluginNativeClient struct {
	connection CallsRPC
	pluginType plugin.PluginType
	encoder    encoding.Encoder
	encrypter  *encrypter.Encrypter
}

func NewCollectorNativeClient(address string, timeout time.Duration, pub *rsa.PublicKey, priv *rsa.PrivateKey) (PluginCollectorClient, error) {
	return newNativeClient(address, timeout, plugin.CollectorPluginType, pub, priv)
}

func NewPublisherNativeClient(address string, timeout time.Duration, pub *rsa.PublicKey, priv *rsa.PrivateKey) (PluginPublisherClient, error) {
	return newNativeClient(address, timeout, plugin.PublisherPluginType, pub, priv)
}

func NewProcessorNativeClient(address string, timeout time.Duration, pub *rsa.PublicKey, priv *rsa.PrivateKey) (PluginProcessorClient, error) {
	return newNativeClient(address, timeout, plugin.ProcessorPluginType, pub, priv)
}

func (p *PluginNativeClient) Ping() error {
	var reply []byte
	err := p.connection.Call("SessionState.Ping", []byte{}, &reply)
	return err
}

func (p *PluginNativeClient) SetKey() error {
	out, err := p.encrypter.EncryptKey()
	if err != nil {
		return err
	}
	return p.connection.Call("SessionState.SetKey", plugin.SetKeyArgs{
		Key: out,
	}, &[]byte{})
}

func (p *PluginNativeClient) Kill(reason string) error {
	args := plugin.KillArgs{Reason: reason}
	out, err := p.encoder.Encode(args)
	if err != nil {
		return err
	}

	var reply []byte
	err = p.connection.Call("SessionState.Kill", out, &reply)
	return err
}

func (p *PluginNativeClient) Publish(contentType string, content []byte, config map[string]ctypes.ConfigValue) error {
	args := plugin.PublishArgs{ContentType: contentType, Content: content, Config: config}

	out, err := p.encoder.Encode(args)
	if err != nil {
		return err
	}

	var reply []byte
	err = p.connection.Call("Publisher.Publish", out, &reply)
	return err
}

func (p *PluginNativeClient) Process(contentType string, content []byte, config map[string]ctypes.ConfigValue) (string, []byte, error) {
	args := plugin.ProcessorArgs{ContentType: contentType, Content: content, Config: config}

	out, err := p.encoder.Encode(args)
	if err != nil {
		return "", nil, err
	}

	var reply []byte
	err = p.connection.Call("Processor.Process", out, &reply)
	if err != nil {
		return "", nil, err
	}

	r := plugin.ProcessorReply{}
	err = p.encoder.Decode(reply, &r)
	if err != nil {
		return "", nil, err
	}

	return r.ContentType, r.Content, nil
}

func (p *PluginNativeClient) CollectMetrics(coreMetricTypes []core.Metric) ([]core.Metric, error) {
	// Convert core.MetricType slice into plugin.PluginMetricType slice as we have
	// to send structs over RPC
	if len(coreMetricTypes) == 0 {
		return nil, errors.New("no metrics to collect")
	}

	var fromCache []core.Metric
	for i, mt := range coreMetricTypes {
		var metric core.Metric
		// Attempt to retreive the metric from the cache. If it is available,
		// nil out that entry in the requested collection.
		if metric = metricCache.get(core.JoinNamespace(mt.Namespace())); metric != nil {
			fromCache = append(fromCache, metric)
			coreMetricTypes[i] = nil
		}
	}
	// If the size of fromCache is equal to the length of the requested metrics,
	// then we retrieved all of the requested metrics and do not need to go
	// through the motions of the rpc call.
	if len(fromCache) != len(coreMetricTypes) {
		var pluginMetricTypes []plugin.PluginMetricType
		// Walk through the requested collection. If the entry is not nil,
		// add it to the slice of metrics to collect over rpc.
		for i, mt := range coreMetricTypes {
			if mt != nil {
				pluginMetricTypes = append(pluginMetricTypes, plugin.PluginMetricType{
					Namespace_:          mt.Namespace(),
					LastAdvertisedTime_: mt.LastAdvertisedTime(),
					Version_:            mt.Version(),
				})
				if mt.Config() != nil {
					pluginMetricTypes[i].Config_ = mt.Config()
				}
			}
		}

		args := plugin.CollectMetricsArgs{PluginMetricTypes: pluginMetricTypes}

		out, err := p.encoder.Encode(args)
		if err != nil {
			return nil, err
		}

		var reply []byte
		err = p.connection.Call("Collector.CollectMetrics", out, &reply)
		if err != nil {
			return nil, err
		}

		r := &plugin.CollectMetricsReply{}
		err = p.encoder.Decode(reply, r)
		if err != nil {
			return nil, err
		}

		var offset int
		for i, mt := range fromCache {
			coreMetricTypes[i] = mt
			offset++
		}
		for i, mt := range r.PluginMetrics {
			metricCache.put(core.JoinNamespace(mt.Namespace_), mt)
			coreMetricTypes[i+offset] = mt
		}
		return coreMetricTypes, err
	}
	return fromCache, nil
}

func (p *PluginNativeClient) GetMetricTypes() ([]core.Metric, error) {
	var reply []byte
	err := p.connection.Call("Collector.GetMetricTypes", []byte{}, &reply)
	if err != nil {
		return nil, err
	}

	r := &plugin.GetMetricTypesReply{}
	err = p.encoder.Decode(reply, r)
	if err != nil {
		return nil, err
	}

	retMetricTypes := make([]core.Metric, len(r.PluginMetricTypes))
	for i, _ := range r.PluginMetricTypes {
		// Set the advertised time
		r.PluginMetricTypes[i].LastAdvertisedTime_ = time.Now()
		retMetricTypes[i] = r.PluginMetricTypes[i]
	}
	return retMetricTypes, nil
}

func (p *PluginNativeClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	var reply []byte
	err := p.connection.Call("SessionState.GetConfigPolicy", []byte{}, &reply)
	if err != nil {
		return nil, err
	}

	r := &plugin.GetConfigPolicyReply{}
	err = p.encoder.Decode(reply, r)
	if err != nil {
		return nil, err
	}

	return r.Policy, nil
}

// GetType returns the string type of the plugin
// Note: the first letter of the type will be capitalized.
func (p *PluginNativeClient) GetType() string {
	return upcaseInitial(p.pluginType.String())
}

func newNativeClient(address string, timeout time.Duration, t plugin.PluginType, pub *rsa.PublicKey, priv *rsa.PrivateKey) (*PluginNativeClient, error) {
	// Attempt to dial address error on timeout or problem
	conn, err := net.DialTimeout("tcp", address, timeout)
	// Return nil RPCClient and err if encoutered
	if err != nil {
		return nil, err
	}
	r := rpc.NewClient(conn)
	p := &PluginNativeClient{
		connection: r,
		pluginType: t,
	}

	key, err := encrypter.GenerateKey()
	if err != nil {
		return nil, err
	}
	p.encoder = encoding.NewGobEncoder()

	encrypter := encrypter.New(pub, priv)
	encrypter.Key = key
	p.encrypter = encrypter
	p.encoder.SetEncrypter(encrypter)

	return p, nil
}

func init() {
	gob.Register(*(&ctypes.ConfigValueStr{}))
	gob.Register(*(&ctypes.ConfigValueInt{}))
	gob.Register(*(&ctypes.ConfigValueFloat{}))

	gob.Register(cpolicy.NewPolicyNode())
	gob.Register(&cpolicy.StringRule{})
	gob.Register(&cpolicy.IntRule{})
	gob.Register(&cpolicy.FloatRule{})
}

func upcaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}
