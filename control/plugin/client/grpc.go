/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encrypter"
	"github.com/intelsdi-x/snap/control/plugin/rpc"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/pkg/rpcutil"
	"google.golang.org/grpc/metadata"
)

// SecureSide identifies security mode to apply in securing gRPC
type SecureSide int

// Define secuity modes available to apply to gRPC.
const (
	SecureClient = SecureSide(iota)
	SecureServer
	DisabledSecurity
)

type pluginClient interface {
	Ping(ctx context.Context, in *rpc.Empty, opts ...grpc.CallOption) (*rpc.ErrReply, error)
	Kill(ctx context.Context, in *rpc.KillArg, opts ...grpc.CallOption) (*rpc.ErrReply, error)
	GetConfigPolicy(ctx context.Context, in *rpc.Empty, opts ...grpc.CallOption) (*rpc.GetConfigPolicyReply, error)
}

type grpcClient struct {
	collector       rpc.CollectorClient
	streamCollector rpc.StreamCollectorClient
	processor       rpc.ProcessorClient
	publisher       rpc.PublisherClient
	plugin          pluginClient
	context         context.Context

	// Channel used to signal death of stream collector to scheduler
	killChan chan struct{}
	// stream connection to stream collector
	stream rpc.StreamCollector_StreamMetricsClient

	pluginType plugin.PluginType
	timeout    time.Duration
	conn       *grpc.ClientConn
	encrypter  *encrypter.Encrypter
}

// GRPCSecurity contains data necessary to setup secure gRPC communication
type GRPCSecurity struct {
	TLSEnabled  bool
	SecureSide  SecureSide
	TLSCertPath string
	TLSKeyPath  string
	CACertPaths []string
}

// SecurityTLSEnabled generates security object for securing gRPC communication
func SecurityTLSEnabled(certPath, keyPath string, secureSide SecureSide) GRPCSecurity {
	return GRPCSecurity{
		TLSEnabled:  true,
		SecureSide:  secureSide,
		TLSCertPath: certPath,
		TLSKeyPath:  keyPath,
	}
}

// SecurityTLSExtended generates security object for securing gRPC communication.
// This function accepts also a list of CA cert paths for verifying TLS participant's
// identity.
func SecurityTLSExtended(certPath, keyPath string, secureSide SecureSide, caCertPaths []string) GRPCSecurity {
	return GRPCSecurity{
		TLSEnabled:  true,
		SecureSide:  secureSide,
		TLSCertPath: certPath,
		TLSKeyPath:  keyPath,
		CACertPaths: caCertPaths,
	}
}

// SecurityTLSOff generates setup object deactivating gRPC security
func SecurityTLSOff() GRPCSecurity {
	return GRPCSecurity{
		TLSEnabled: false,
		SecureSide: DisabledSecurity,
	}
}

// NewCollectorGrpcClient returns a collector gRPC Client.
func NewCollectorGrpcClient(address string, timeout time.Duration, security GRPCSecurity) (PluginCollectorClient, error) {
	ctx := context.Background()
	p, err := newPluginGrpcClient(ctx, address, timeout, security, plugin.CollectorPluginType)
	if err != nil {
		return nil, err
	}
	return p.(PluginCollectorClient), err
}

// NewStreamCollectorGrpcClient returns a stream collector gRPC client
func NewStreamCollectorGrpcClient(address string, timeout time.Duration, security GRPCSecurity) (PluginStreamCollectorClient, error) {
	ctx := context.Background()
	p, err := newPluginGrpcClient(ctx, address, timeout, security, plugin.StreamCollectorPluginType)
	if err != nil {
		return nil, err
	}
	return p.(PluginStreamCollectorClient), nil
}

// NewProcessorGrpcClient returns a processor gRPC Client.
func NewProcessorGrpcClient(address string, timeout time.Duration, security GRPCSecurity) (PluginProcessorClient, error) {
	ctx := context.Background()
	p, err := newPluginGrpcClient(ctx, address, timeout, security, plugin.ProcessorPluginType)
	if err != nil {
		return nil, err
	}
	return p.(PluginProcessorClient), err
}

// NewPublisherGrpcClient returns a publisher gRPC Client.
func NewPublisherGrpcClient(address string, timeout time.Duration, security GRPCSecurity) (PluginPublisherClient, error) {
	ctx := context.Background()
	p, err := newPluginGrpcClient(ctx, address, timeout, security, plugin.PublisherPluginType)
	if err != nil {
		return nil, err
	}
	return p.(PluginPublisherClient), err
}

func loadCACerts(certPaths []string) (rootCAs *x509.CertPool, err error) {
	var path string
	var filepaths []string
	// list potential certificate files
	for _, path := range certPaths {
		var stat os.FileInfo
		if stat, err = os.Stat(path); err != nil {
			return nil, fmt.Errorf("unable to process CA cert source path %s: %v", path, err)
		}
		if !stat.IsDir() {
			filepaths = append(filepaths, path)
			continue
		}
		var subfiles []os.FileInfo
		if subfiles, err = ioutil.ReadDir(path); err != nil {
			return nil, fmt.Errorf("unable to process CA cert source directory %s: %v", path, err)
		}
		for _, subfile := range subfiles {
			subpath := filepath.Join(path, subfile.Name())
			if subfile.IsDir() {
				log.WithField("path", subpath).Debug("Skipping second level directory found among certificate files")
				continue
			}
			filepaths = append(filepaths, subpath)
		}
	}
	rootCAs = x509.NewCertPool()
	numread := 0
	for _, path = range filepaths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.WithFields(log.Fields{"path": path, "error": err}).Debug("Unable to read cert file")
			continue
		}
		if !rootCAs.AppendCertsFromPEM(b) {
			log.WithField("path", path).Debug("Didn't find any usable certificates in cert file")
			continue
		}
		numread++
	}
	if numread == 0 {
		return nil, fmt.Errorf("found no usable certificates in given locations")
	}
	return rootCAs, nil
}

func buildCredentials(security GRPCSecurity) (creds credentials.TransportCredentials, err error) {
	if !security.TLSEnabled {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(security.TLSCertPath, security.TLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to load TLS key pair: %v", err)
	}
	var rootCAs *x509.CertPool
	if len(security.CACertPaths) > 0 {
		log.Debug("Loading CA certificates given explicitly")
		rootCAs, err = loadCACerts(security.CACertPaths)
		if err != nil {
			return nil, err
		}
	} else {
		log.Debug("Loading CA certificates from operating system")
		rootCAs, err = x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("unable to load system-wide root TLS certificates: %v", err)
		}
	}
	switch security.SecureSide {
	case SecureClient:
		creds = credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      rootCAs,
		})
	case SecureServer:
		creds = credentials.NewTLS(&tls.Config{
			Certificates:             []tls.Certificate{cert},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			},
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  rootCAs,
		})
	case DisabledSecurity:
		creds = nil
	}
	return creds, nil
}

// newPluginGrpcClient returns a configured gRPC Client.
func newPluginGrpcClient(ctx context.Context, address string, timeout time.Duration, security GRPCSecurity, typ plugin.PluginType) (interface{}, error) {
	address, port, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	var p *grpcClient
	var creds credentials.TransportCredentials
	if creds, err = buildCredentials(security); err != nil {
		return nil, err
	}
	p, err = newGrpcClient(ctx, address, int(port), timeout, typ, creds)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func parseAddress(address string) (string, int64, error) {
	addr := strings.Split(address, ":")
	if len(addr) != 2 {
		return "", 0, fmt.Errorf("bad address")
	}
	address = addr[0]
	port, err := strconv.ParseInt(addr[1], 10, 64)
	if err != nil {
		return "", 0, err
	}
	return address, port, nil
}

func newGrpcClient(ctx context.Context, addr string, port int, timeout time.Duration, typ plugin.PluginType, creds credentials.TransportCredentials) (*grpcClient, error) {
	var conn *grpc.ClientConn
	var err error
	if conn, err = rpcutil.GetClientConnectionWithCreds(ctx, addr, port, creds); err != nil {
		return nil, err
	}
	p := &grpcClient{
		timeout: timeout,
		conn:    conn,
		context: ctx,
	}

	switch typ {
	case plugin.CollectorPluginType:
		p.collector = rpc.NewCollectorClient(conn)
		p.plugin = p.collector
	case plugin.StreamCollectorPluginType:
		p.streamCollector = rpc.NewStreamCollectorClient(conn)
		p.plugin = p.streamCollector
		p.killChan = make(chan struct{})
	case plugin.ProcessorPluginType:
		p.processor = rpc.NewProcessorClient(conn)
		p.plugin = p.processor
	case plugin.PublisherPluginType:
		p.publisher = rpc.NewPublisherClient(conn)
		p.plugin = p.publisher
	default:
		return nil, errors.New(fmt.Sprintf("Invalid plugin type provided %v", typ))
	}

	return p, nil
}

func getContext(timeout time.Duration) context.Context {
	ctxTimeout, _ := context.WithTimeout(context.Background(), timeout)
	return ctxTimeout
}

func (g *grpcClient) Ping() error {
	_, err := g.plugin.Ping(getContext(g.timeout), &rpc.Empty{})
	if err != nil {
		return err
	}
	return nil
}

func (g *grpcClient) SetKey() error {
	// Added to conform to interface but not needed by grpc
	return nil
}

// Killed closes the killChan for a streaming rpc
func (g *grpcClient) Killed() {
	if g.killChan != nil {
		close(g.killChan)
	}
}

func (g *grpcClient) Kill(reason string) error {
	_, err := g.plugin.Kill(getContext(g.timeout), &rpc.KillArg{Reason: reason})
	g.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (g *grpcClient) Close() error {
	err := g.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (g *grpcClient) Publish(metrics []core.Metric, config map[string]ctypes.ConfigValue) error {
	arg := &rpc.PubProcArg{
		Metrics: NewMetrics(metrics),
		Config:  ToConfigMap(config),
	}
	reply, err := g.publisher.Publish(getContext(g.timeout), arg)
	if err != nil {
		return err
	}
	if reply.Error != "" {
		return errors.New(reply.Error)
	}
	return nil
}

func (g *grpcClient) Process(metrics []core.Metric, config map[string]ctypes.ConfigValue) ([]core.Metric, error) {
	arg := &rpc.PubProcArg{
		Metrics: NewMetrics(metrics),
		Config:  ToConfigMap(config),
	}
	reply, err := g.processor.Process(getContext(g.timeout), arg)

	if err != nil {
		return nil, err
	}
	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}
	mts := ToCoreMetrics(reply.Metrics)
	for _, mt := range mts {
		log.Debug(mt.Namespace())
	}
	return mts, nil
}

func (g *grpcClient) CollectMetrics(mts []core.Metric) ([]core.Metric, error) {
	arg := &rpc.MetricsArg{
		Metrics: NewMetrics(mts),
	}
	reply, err := g.collector.CollectMetrics(getContext(g.timeout), arg)

	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	metrics := ToCoreMetrics(reply.Metrics)
	return metrics, nil
}

func (g *grpcClient) UpdateCollectedMetrics(mts []core.Metric) error {
	if g.stream != nil {
		arg := &rpc.CollectArg{
			Metrics_Arg: &rpc.MetricsArg{Metrics: NewMetrics(mts)},
		}
		err := g.stream.Send(arg)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *grpcClient) UpdatePluginConfig(bytes []byte) error {
	if g.stream != nil {
		arg := &rpc.CollectArg{
			Other: bytes,
		}
		err := g.stream.Send(arg)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *grpcClient) UpdateCollectDuration(maxCollectDuration time.Duration) error {
	if g.stream != nil {
		arg := &rpc.CollectArg{
			MaxCollectDuration: maxCollectDuration.Nanoseconds(),
		}
		err := g.stream.Send(arg)
		if err != nil {
			return err
		}
	}
	return nil
}
func (g *grpcClient) UpdateMetricsBuffer(maxMetricsBuffer int64) error {
	if g.stream != nil {
		arg := &rpc.CollectArg{
			MaxMetricsBuffer: maxMetricsBuffer,
		}
		err := g.stream.Send(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *grpcClient) StreamMetrics(taskID string, mts []core.Metric) (chan []core.Metric, chan error, error) {
	arg := &rpc.CollectArg{
		Metrics_Arg: &rpc.MetricsArg{Metrics: NewMetrics(mts)},
	}
	if len(mts) == 0 {
		return nil, nil, errors.New("No metrics requested to stream")
	}

	mt := mts[0]
	if cfg := mt.Config(); cfg != nil {
		values := cfg.Table()
		if values != nil {
			maxCollectDuration, ok := values["MaxCollectDuration"]
			if ok {
				t, ok := maxCollectDuration.(ctypes.ConfigValueInt)
				if ok {
					// MaxCollectDuration was passed as an int therefore
					// it is representing nanoseconds
					arg.MaxCollectDuration = int64(t.Value)
				} else {
					t, ok := maxCollectDuration.(ctypes.ConfigValueStr)
					if ok {
						// MaxCollectDuration was passed as a string therefore
						// it should be a string rep of a duration
						dur, err := time.ParseDuration(t.Value)
						if err != nil {
							log.WithFields(
								log.Fields{
									"_block":     "StreamMetrics",
									"config-key": "MaxCollectDuration",
									"hint":       "value should be a parsable duration (e.g. 5s)",
									"error":      err.Error(),
								},
							).Warn("invalid config value")
						}
						arg.MaxCollectDuration = dur.Nanoseconds()
					} else {
						log.WithFields(
							log.Fields{
								"_block":        "StreamMetrics",
								"config-key":    "MaxCollectDuration",
								"type-provided": reflect.TypeOf(maxCollectDuration).String(),
								"type-wanted":   ctypes.ConfigValueStr{}.Type(),
								"hint":          "value should be a parsable duration (e.g. 5s)",
							},
						).Warn("wrong config value type")
					}
				}
			}
			maxMetricsBuffer, ok := values["MaxMetricsBuffer"]
			if ok {
				t, ok := maxMetricsBuffer.(ctypes.ConfigValueInt)
				if ok {
					arg.MaxMetricsBuffer = int64(t.Value)
				}
			}
		}
	}

	header := metadata.New(map[string]string{
		"task-id": taskID,
	})
	ctx := metadata.NewContext(g.context, header)

	s, err := g.streamCollector.StreamMetrics(ctx)
	if err != nil {
		return nil, nil, err
	}
	err = s.Send(arg)
	if err != nil {
		return nil, nil, err
	}
	metricChan := make(chan []core.Metric)
	errChan := make(chan error)
	doneChan := make(chan struct{})
	g.killChan = doneChan
	g.stream = s
	go g.handleInStream(metricChan, errChan)
	return metricChan, errChan, nil
}

func (g *grpcClient) handleInStream(
	metricChan chan []core.Metric,
	errChan chan error) {
	go func() {
		for {
			in, err := g.stream.Recv()
			if err != nil {
				g.conn.Close()
				if strings.Contains(err.Error(), "transport is closing") {
					errChan <- errors.New("connection broken")
				}
				errChan <- err
				break
			}
			if in.Metrics_Reply != nil {
				mts := ToCoreMetrics(in.Metrics_Reply.Metrics)
				if len(mts) == 0 {
					// skip empty metrics
					continue
				}
				metricChan <- mts
			} else if in.Error != nil {
				e := errors.New(in.Error.Error)
				errChan <- e
			}
		}
	}()

	<-g.killChan
	errChan <- errors.New("connection broken")

}

func (g *grpcClient) GetMetricTypes(config plugin.ConfigType) ([]core.Metric, error) {
	arg := &rpc.GetMetricTypesArg{
		Config: ToConfigMap(config.Table()),
	}
	var reply *rpc.MetricsReply
	var err error
	if g.streamCollector != nil {
		reply, err = g.streamCollector.GetMetricTypes(getContext(g.timeout), arg)
	} else {
		reply, err = g.collector.GetMetricTypes(getContext(g.timeout), arg)
	}
	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	results := ToCoreMetrics(reply.Metrics)
	return results, nil
}

func (g *grpcClient) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	reply, err := g.plugin.GetConfigPolicy(getContext(g.timeout), &rpc.Empty{})

	if err != nil {
		return nil, err
	}

	if reply.Error != "" {
		return nil, errors.New(reply.Error)
	}

	return rpc.ToConfigPolicy(reply), nil
}

type metric struct {
	namespace          core.Namespace
	version            int
	config             *cdata.ConfigDataNode
	lastAdvertisedTime time.Time
	timeStamp          time.Time
	data               interface{}
	tags               map[string]string
	description        string
	unit               string
}

func (m *metric) Namespace() core.Namespace     { return m.namespace }
func (m *metric) Config() *cdata.ConfigDataNode { return m.config }
func (m *metric) Version() int                  { return m.version }
func (m *metric) Data() interface{}             { return m.data }
func (m *metric) Tags() map[string]string       { return m.tags }
func (m *metric) LastAdvertisedTime() time.Time { return m.lastAdvertisedTime }
func (m *metric) Timestamp() time.Time          { return m.timeStamp }
func (m *metric) Description() string           { return m.description }
func (m *metric) Unit() string                  { return m.unit }

func ToCoreMetrics(mts []*rpc.Metric) []core.Metric {
	metrics := make([]core.Metric, len(mts))
	for i, mt := range mts {
		metrics[i] = ToCoreMetric(mt)
	}
	return metrics
}

func ToCoreMetric(mt *rpc.Metric) core.Metric {
	if mt.Timestamp == nil {
		mt.Timestamp = ToTime(time.Now())
	}

	if mt.LastAdvertisedTime == nil {
		mt.LastAdvertisedTime = ToTime(time.Now())
	}

	ret := &metric{
		namespace:          ToCoreNamespace(mt.Namespace),
		version:            int(mt.Version),
		tags:               mt.Tags,
		timeStamp:          time.Unix(mt.Timestamp.Sec, mt.Timestamp.Nsec),
		lastAdvertisedTime: time.Unix(mt.LastAdvertisedTime.Sec, mt.LastAdvertisedTime.Nsec),
		config:             ConfigMapToConfig(mt.Config),
		description:        mt.Description,
		unit:               mt.Unit,
	}

	switch mt.Data.(type) {
	case *rpc.Metric_BytesData:
		ret.data = mt.GetBytesData()
	case *rpc.Metric_StringData:
		ret.data = mt.GetStringData()
	case *rpc.Metric_Float32Data:
		ret.data = mt.GetFloat32Data()
	case *rpc.Metric_Float64Data:
		ret.data = mt.GetFloat64Data()
	case *rpc.Metric_Int32Data:
		ret.data = mt.GetInt32Data()
	case *rpc.Metric_Int64Data:
		ret.data = mt.GetInt64Data()
	case *rpc.Metric_BoolData:
		ret.data = mt.GetBoolData()
	case *rpc.Metric_Uint32Data:
		ret.data = mt.GetUint32Data()
	case *rpc.Metric_Uint64Data:
		ret.data = mt.GetUint64Data()
	}
	return ret
}

func NewMetrics(ms []core.Metric) []*rpc.Metric {
	metrics := make([]*rpc.Metric, len(ms))
	for i, m := range ms {
		metrics[i] = ToMetric(m)
	}
	return metrics
}

func ToMetric(co core.Metric) *rpc.Metric {
	cm := &rpc.Metric{
		Namespace: ToNamespace(co.Namespace()),
		Version:   int64(co.Version()),
		Tags:      co.Tags(),
		Timestamp: &rpc.Time{
			Sec:  co.Timestamp().Unix(),
			Nsec: int64(co.Timestamp().Nanosecond()),
		},
		LastAdvertisedTime: &rpc.Time{
			Sec:  co.LastAdvertisedTime().Unix(),
			Nsec: int64(co.Timestamp().Nanosecond()),
		},
		Unit: co.Unit(),
	}
	if co.Config() != nil {
		cm.Config = ConfigToConfigMap(co.Config())
	}
	switch t := co.Data().(type) {
	case string:
		cm.Data = &rpc.Metric_StringData{t}
	case float64:
		cm.Data = &rpc.Metric_Float64Data{t}
	case float32:
		cm.Data = &rpc.Metric_Float32Data{t}
	case int32:
		cm.Data = &rpc.Metric_Int32Data{t}
	case int:
		cm.Data = &rpc.Metric_Int64Data{int64(t)}
	case int64:
		cm.Data = &rpc.Metric_Int64Data{t}
	case uint32:
		cm.Data = &rpc.Metric_Uint32Data{t}
	case uint64:
		cm.Data = &rpc.Metric_Uint64Data{t}
	case []byte:
		cm.Data = &rpc.Metric_BytesData{t}
	case bool:
		cm.Data = &rpc.Metric_BoolData{t}
	case nil:
		cm.Data = nil
	default:
		log.Error(fmt.Sprintf("unsupported type: %s", t))
	}
	return cm
}

func ToCoreNamespace(n []*rpc.NamespaceElement) core.Namespace {
	var namespace core.Namespace
	for _, val := range n {
		ele := core.NamespaceElement{
			Value:       val.Value,
			Description: val.Description,
			Name:        val.Name,
		}
		namespace = append(namespace, ele)
	}
	return namespace
}

func ConfigMapToConfig(cfg *rpc.ConfigMap) *cdata.ConfigDataNode {
	if cfg == nil {
		return nil
	}
	config := cdata.FromTable(ParseConfig(cfg))
	return config
}

func ToConfigMap(cv map[string]ctypes.ConfigValue) *rpc.ConfigMap {
	newConfig := &rpc.ConfigMap{
		IntMap:    make(map[string]int64),
		FloatMap:  make(map[string]float64),
		StringMap: make(map[string]string),
		BoolMap:   make(map[string]bool),
	}
	for k, v := range cv {
		switch v.Type() {
		case "integer":
			newConfig.IntMap[k] = int64(v.(ctypes.ConfigValueInt).Value)
		case "float":
			newConfig.FloatMap[k] = v.(ctypes.ConfigValueFloat).Value
		case "string":
			newConfig.StringMap[k] = v.(ctypes.ConfigValueStr).Value
		case "bool":
			newConfig.BoolMap[k] = v.(ctypes.ConfigValueBool).Value
		}
	}
	return newConfig
}

func ToNamespace(n core.Namespace) []*rpc.NamespaceElement {
	elements := make([]*rpc.NamespaceElement, 0, len(n))
	for _, value := range n {
		ne := &rpc.NamespaceElement{
			Value:       value.Value,
			Description: value.Description,
			Name:        value.Name,
		}
		elements = append(elements, ne)
	}
	return elements
}

func ConfigToConfigMap(cd *cdata.ConfigDataNode) *rpc.ConfigMap {
	if cd == nil {
		return nil
	}
	return ToConfigMap(cd.Table())
}

func ParseConfig(config *rpc.ConfigMap) map[string]ctypes.ConfigValue {
	c := make(map[string]ctypes.ConfigValue)
	for k, v := range config.IntMap {
		ival := ctypes.ConfigValueInt{Value: int(v)}
		c[k] = ival
	}
	for k, v := range config.FloatMap {
		fval := ctypes.ConfigValueFloat{Value: v}
		c[k] = fval
	}
	for k, v := range config.StringMap {
		sval := ctypes.ConfigValueStr{Value: v}
		c[k] = sval
	}
	for k, v := range config.BoolMap {
		bval := ctypes.ConfigValueBool{Value: v}
		c[k] = bval
	}
	return c
}

func ToTime(t time.Time) *rpc.Time {
	return &rpc.Time{
		Nsec: t.Unix(),
		Sec:  int64(t.Second()),
	}
}
