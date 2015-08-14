package client

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

type mockProxy struct {
}

func (m *mockProxy) CollectMetrics(args plugin.CollectMetricsArgs, reply *plugin.CollectMetricsReply) error {
	rand.Seed(time.Now().Unix())
	for _, i := range args.PluginMetricTypes {
		p := plugin.NewPluginMetricType(i.Namespace(), rand.Intn(100))
		p.Config_ = i.Config()
		reply.PluginMetrics = append(reply.PluginMetrics, *p)
	}
	return nil
}

func (m *mockProxy) GetMetricTypes(args plugin.GetMetricTypesArgs, reply *plugin.GetMetricTypesReply) error {
	pmts := []plugin.PluginMetricType{}
	pmts = append(pmts, plugin.PluginMetricType{
		Namespace_: []string{"foo", "bar"},
	})
	reply.PluginMetricTypes = pmts
	return nil
}

func (m *mockProxy) GetConfigPolicyTree(args plugin.GetConfigPolicyTreeArgs, reply *plugin.GetConfigPolicyTreeReply) error {
	cpt := cpolicy.NewTree()
	n1 := cpolicy.NewPolicyNode()
	r1, _ := cpolicy.NewStringRule("name", false, "bob")
	n1.Add(r1)
	r2, _ := cpolicy.NewIntegerRule("someInt", true, 100)
	n1.Add(r2)
	r3, _ := cpolicy.NewStringRule("password", true)
	n1.Add(r3)
	r4, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
	n1.Add(r4)
	cpt.Add([]string{"foo", "bar"}, n1)
	reply.PolicyTree = *cpt
	return nil
}

func (m *mockProxy) GetConfigPolicyNode(arg plugin.GetConfigPolicyNodeArgs, reply *plugin.GetConfigPolicyNodeReply) error {
	cpn := cpolicy.NewPolicyNode()
	r1, _ := cpolicy.NewIntegerRule("SomeRequiredInt", true, 1)
	r2, _ := cpolicy.NewStringRule("password", true)
	r3, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
	cpn.Add(r1, r2, r3)
	reply.PolicyNode = *cpn
	return nil
}

func (m *mockProxy) Process(args plugin.ProcessorArgs, reply *plugin.ProcessorReply) error {
	reply.Content = args.Content
	reply.ContentType = args.ContentType
	return nil
}

func (m *mockProxy) Publish(args plugin.PublishArgs, reply *plugin.PublishReply) error {
	return nil
}

type mockSessionStatePluginProxy struct {
}

func (m *mockSessionStatePluginProxy) Ping(arg plugin.PingArgs, b *bool) error {
	*b = true
	return nil
}

func (m *mockSessionStatePluginProxy) Kill(arg plugin.KillArgs, b *bool) error {
	*b = true
	return nil
}

var (
	PluginName = "pulse-collector-dummy2"
	PulsePath  = os.Getenv("PULSE_PATH")
	PluginPath = path.Join(PulsePath, "plugin", PluginName)
)

var httpStarted = false

func startHTTPJSONRPC() string {
	mockProxy := &mockProxy{}
	rpc.RegisterName("Collector", mockProxy)
	rpc.RegisterName("Processor", mockProxy)
	rpc.RegisterName("Publisher", mockProxy)
	session := &mockSessionStatePluginProxy{}
	rpc.RegisterName("SessionState", session)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			res := plugin.NewRPCRequest(req.Body).Call()
			io.Copy(w, res)
		})
		http.Serve(l, nil)
	}()

	return l.Addr().String()
}

func TestHTTPJSONRPC(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr := startHTTPJSONRPC()
	time.Sleep(time.Millisecond * 100)

	Convey("JSON RPC over http", t, func() {
		So(addr, ShouldNotEqual, "")

		Convey("call", func() {
			client := &httpJSONRPCClient{
				url: fmt.Sprintf("http://%v/rpc", addr),
			}

			Convey("method = SessionState.Ping", func() {
				result, err := client.call("SessionState.Ping", []interface{}{plugin.PingArgs{}})
				So(err, ShouldBeNil)
				So(result, ShouldNotResemble, "")
				So(result["result"], ShouldEqual, true)
			})

			Convey("method = Collector.CollectMetrics", func() {
				req := plugin.PluginMetricType{Namespace_: []string{"foo", "bar"}}
				result, err := client.call("Collector.CollectMetrics", []interface{}{[]core.Metric{req}})
				So(err, ShouldBeNil)
				So(result, ShouldNotResemble, "")
				So(result["result"], ShouldHaveSameTypeAs, map[string]interface{}{})
			})

			Convey("method = Collector.GetMetricTypes", func() {
				result, err := client.call("Collector.GetMetricTypes", []interface{}{})
				So(err, ShouldBeNil)
				So(result, ShouldNotResemble, "")
				So(result["result"], ShouldHaveSameTypeAs, map[string]interface{}{})
			})

			Convey("method = Collector.GetConfigPolicyTree", func() {
				result, err := client.call("Collector.GetConfigPolicyTree", []interface{}{})
				So(err, ShouldBeNil)
				So(result, ShouldNotResemble, "")
				So(result["result"], ShouldHaveSameTypeAs, map[string]interface{}{})
			})

			Convey("method = Processor.GetConfigPolicyNode", func() {
				result, err := client.call("Processor.GetConfigPolicyNode", []interface{}{})
				So(err, ShouldBeNil)
				So(result, ShouldNotResemble, "")
				So(result["result"], ShouldHaveSameTypeAs, map[string]interface{}{})
			})

			Convey("method = Processor.Process", func() {
				result, err := client.call("Processor.Process", []interface{}{})
				So(err, ShouldBeNil)
				So(result, ShouldNotResemble, "")
				So(result["result"], ShouldHaveSameTypeAs, map[string]interface{}{})
			})

			Convey("method = Publisher.Publish", func() {
				_, err := client.call("Publisher.Publish", []interface{}{})
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Collector Client", t, func() {
		c := NewCollectorHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second)
		So(c, ShouldNotBeNil)

		Convey("Ping", func() {
			err := c.Ping()
			So(err, ShouldBeNil)
		})

		Convey("Kill", func() {
			err := c.Kill("somereason")
			So(err, ShouldBeNil)
		})

		Convey("GetMetricTypes", func() {
			mts, err := c.GetMetricTypes()
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
			So(mts, ShouldHaveSameTypeAs, []core.Metric{})
			So(len(mts), ShouldBeGreaterThan, 0)
		})

		Convey("CollectMetrics provided a valid config", func() {
			cdn := cdata.NewNode()
			cdn.AddItem("someInt", ctypes.ConfigValueInt{Value: 1})
			cdn.AddItem("password", ctypes.ConfigValueStr{Value: "secure"})

			mts, err := c.CollectMetrics([]core.Metric{
				&plugin.PluginMetricType{
					Namespace_: []string{"foo", "bar"},
					Config_:    cdn,
				},
			})
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
			So(mts, ShouldHaveSameTypeAs, []core.Metric{})
			So(len(mts), ShouldBeGreaterThan, 0)
			So(mts[0].Config().Table(), ShouldNotBeEmpty)
			So(mts[0].Config().Table()["someInt"].Type(), ShouldResemble, "integer")

			Convey("Get and process the ConfigPolicyTree", func() {
				cpt, err := c.GetConfigPolicyTree()
				So(err, ShouldBeNil)
				So(cpt, ShouldNotBeNil)
				So(cpt.Get([]string{"foo", "bar"}), ShouldNotBeNil)
				node := cpt.Get([]string{"foo", "bar"})
				So(err, ShouldBeNil)
				So(node, ShouldNotBeNil)
				cpn, cperrs := node.Process(mts[0].Config().Table())
				So(cpn, ShouldNotBeNil)
				So((*cpn)["somefloat"].Type(), ShouldResemble, "float")
				So((*cpn)["somefloat"].(*ctypes.ConfigValueFloat).Value, ShouldResemble, 3.14)
				So(cperrs.Errors(), ShouldBeEmpty)
			})
		})

		Convey("CollectMetrics provided an invalid config", func() {
			cdn := cdata.NewNode()
			cdn.AddItem("someInt", ctypes.ConfigValueInt{Value: 1})

			mts, err := c.CollectMetrics([]core.Metric{
				&plugin.PluginMetricType{
					Namespace_: []string{"foo", "bar"},
					Config_:    cdn,
				},
			})
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
			So(mts, ShouldHaveSameTypeAs, []core.Metric{})
			So(len(mts), ShouldBeGreaterThan, 0)
			So(mts[0].Config().Table(), ShouldNotBeEmpty)
			So(mts[0].Config().Table()["someInt"].Type(), ShouldResemble, "integer")

			Convey("Get and proces the ConfigPolicyTree", func() {
				cpt, err := c.GetConfigPolicyTree()
				So(err, ShouldBeNil)
				So(cpt, ShouldNotBeNil)
				So(cpt.Get([]string{"foo", "bar"}), ShouldNotBeNil)
				node := cpt.Get([]string{"foo", "bar"})
				So(err, ShouldBeNil)
				So(node, ShouldNotBeNil)
				cpn, cperrs := node.Process(mts[0].Config().Table())
				So(cpn, ShouldBeNil)
				So(cperrs.Errors(), ShouldNotBeEmpty)
				So(len(cperrs.Errors()), ShouldEqual, 1)
				So(cperrs.Errors()[0].Error(), ShouldContainSubstring, "password")
			})
		})

		Convey("Processor Client", func() {
			p := NewProcessorHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second)
			So(c, ShouldNotBeNil)

			Convey("Ping", func() {
				err := p.Ping()
				So(err, ShouldBeNil)
			})

			Convey("Kill", func() {
				err := p.Kill("somereason")
				So(err, ShouldBeNil)
			})

			Convey("GetConfigPolicyNode", func() {
				cpn, err := p.GetConfigPolicyNode()
				So(err, ShouldBeNil)
				So(cpn, ShouldNotBeNil)
				cpn_ := cpolicy.NewPolicyNode()
				r1, err := cpolicy.NewIntegerRule("SomeRequiredInt", true, 1)
				r2, _ := cpolicy.NewStringRule("password", true)
				r3, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
				So(err, ShouldBeNil)
				cpn_.Add(r1, r2, r3)
				cpnjson, _ := cpn.MarshalJSON()
				cpn_json, _ := cpn_.MarshalJSON()
				So(string(cpnjson), ShouldResemble, string(cpn_json))
			})

			Convey("Process metrics", func() {
				pmt := plugin.NewPluginMetricType([]string{"foo", "bar"}, 1)
				b, _ := json.Marshal([]plugin.PluginMetricType{*pmt})
				contentType, content, err := p.Process(plugin.PulseJSONContentType, b, nil)
				So(contentType, ShouldResemble, plugin.PulseJSONContentType)
				So(content, ShouldNotBeNil)
				So(err, ShouldEqual, nil)
				var pmts []plugin.PluginMetricType
				err = json.Unmarshal(content, &pmts)
				So(err, ShouldBeNil)
				So(len(pmts), ShouldEqual, 1)
				So(pmts[0].Data(), ShouldEqual, 1)
				So(pmts[0].Namespace(), ShouldResemble, []string{"foo", "bar"})
			})
		})

		Convey("Publisher Client", func() {
			p := NewPublisherHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second)
			So(c, ShouldNotBeNil)

			Convey("Ping", func() {
				err := p.Ping()
				So(err, ShouldBeNil)
			})

			Convey("Kill", func() {
				err := p.Kill("somereason")
				So(err, ShouldBeNil)
			})

			Convey("GetConfigPolicyNode", func() {
				cpn, err := p.GetConfigPolicyNode()
				So(err, ShouldBeNil)
				So(cpn, ShouldNotBeNil)
				cpn_ := cpolicy.NewPolicyNode()
				r1, err := cpolicy.NewIntegerRule("SomeRequiredInt", true, 1)
				r2, _ := cpolicy.NewStringRule("password", true)
				r3, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
				So(err, ShouldBeNil)
				cpn_.Add(r1, r2, r3)
				cpnjson, _ := cpn.MarshalJSON()
				cpn_json, _ := cpn_.MarshalJSON()
				So(string(cpnjson), ShouldResemble, string(cpn_json))
			})

			Convey("Publish metrics", func() {
				pmt := plugin.NewPluginMetricType([]string{"foo", "bar"}, 1)
				b, _ := json.Marshal([]plugin.PluginMetricType{*pmt})
				err := p.Publish(plugin.PulseJSONContentType, b, nil)
				So(err, ShouldBeNil)
			})

		})

	})
}
