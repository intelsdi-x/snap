/*          **  DEPRECATED  **
For more information, see our deprecation notice
on Github: https://github.com/intelsdi-x/snap/issues/1296
*/

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
	crand "crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/control/plugin/encoding"
	"github.com/intelsdi-x/snap/control/plugin/encrypter"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	key, _    = rsa.GenerateKey(crand.Reader, 1024)
	symkey, _ = encrypter.GenerateKey()
)

type mockProxy struct {
	e encoding.Encoder
}

func (m *mockProxy) Process(args []byte, reply *[]byte) error {
	var dargs plugin.ProcessorArgs
	m.e.Decode(args, &dargs)
	pr := plugin.ProcessorReply{Content: dargs.Content, ContentType: dargs.ContentType}
	*reply, _ = m.e.Encode(pr)
	return nil
}

func (m *mockProxy) Publish(args []byte, reply *[]byte) error {
	return nil
}

type mockCollectorProxy struct {
	e encoding.Encoder
}

func (m *mockCollectorProxy) CollectMetrics(args []byte, reply *[]byte) error {
	rand.Seed(time.Now().Unix())
	var dargs plugin.CollectMetricsArgs
	err := m.e.Decode(args, &dargs)
	if err != nil {
		return err
	}
	var mts []plugin.MetricType
	for _, i := range dargs.MetricTypes {
		p := plugin.NewMetricType(i.Namespace(), time.Now(), nil, "", rand.Intn(100))
		p.Config_ = i.Config()
		mts = append(mts, *p)
	}
	cmr := &plugin.CollectMetricsReply{PluginMetrics: mts}
	*reply, err = m.e.Encode(cmr)
	if err != nil {
		return err
	}
	return nil
}

func (m *mockCollectorProxy) GetMetricTypes(args []byte, reply *[]byte) error {
	dargs := &plugin.GetMetricTypesArgs{}
	m.e.Decode(args, &dargs)

	pmts := []plugin.MetricType{}
	pmts = append(pmts, plugin.MetricType{
		Namespace_: core.NewNamespace("foo", "bar"),
		Config_:    dargs.PluginConfig.ConfigDataNode,
	})
	*reply, _ = m.e.Encode(plugin.GetMetricTypesReply{MetricTypes: pmts})
	return nil
}

type mockSessionStatePluginProxy struct {
	e encoding.Encoder
	c bool
}

func (m *mockSessionStatePluginProxy) GetConfigPolicy(args []byte, reply *[]byte) error {
	cp := cpolicy.New()
	n1 := cpolicy.NewPolicyNode()
	if m.c {
		r1, _ := cpolicy.NewStringRule("name", false, "bob")
		n1.Add(r1)
		r2, _ := cpolicy.NewIntegerRule("someInt", true, 100)
		n1.Add(r2)
		r3, _ := cpolicy.NewStringRule("password", true)
		n1.Add(r3)
		r4, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
		n1.Add(r4)
		cp.Add([]string{"foo", "bar"}, n1)
	} else {
		r1, _ := cpolicy.NewIntegerRule("SomeRequiredInt", true, 1)
		r2, _ := cpolicy.NewStringRule("password", true)
		r3, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
		n1.Add(r1, r2, r3)
		cp.Add([]string{""}, n1)
	}
	cpr := plugin.GetConfigPolicyReply{Policy: cp}
	var err error
	*reply, err = m.e.Encode(cpr)
	return err
}

func (m *mockSessionStatePluginProxy) Ping(arg []byte, b *[]byte) error {
	return nil
}

func (m *mockSessionStatePluginProxy) Kill(arg []byte, b *[]byte) error {
	return nil
}

var httpStarted = false

func startHTTPJSONRPC() (string, *mockSessionStatePluginProxy) {
	encr := encrypter.New(&key.PublicKey, nil)
	encr.Key = symkey
	ee := encoding.NewJsonEncoder()
	ee.SetEncrypter(encr)
	mockProxy := &mockProxy{e: ee}
	mockCollectorProxy := &mockCollectorProxy{e: ee}
	rpc.RegisterName("Collector", mockCollectorProxy)
	rpc.RegisterName("Processor", mockProxy)
	rpc.RegisterName("Publisher", mockProxy)
	session := &mockSessionStatePluginProxy{e: ee}
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

	return l.Addr().String(), session
}

func TestHTTPJSONRPC(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	addr, session := startHTTPJSONRPC()
	time.Sleep(time.Millisecond * 100)

	Convey("Collector Client", t, func() {
		session.c = true
		c, err := NewCollectorHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second, &key.PublicKey, true)
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		cl := c.(*httpJSONRPCClient)
		cl.encrypter.Key = symkey

		Convey("Ping", func() {
			err := c.Ping()
			So(err, ShouldBeNil)
		})

		Convey("Kill", func() {
			err := c.Kill("somereason")
			So(err, ShouldBeNil)
		})

		Convey("GetMetricTypes", func() {
			cfg := plugin.NewPluginConfigType()
			cfg.AddItem("test", ctypes.ConfigValueBool{Value: true})
			mts, err := c.GetMetricTypes(cfg)
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
			So(mts, ShouldHaveSameTypeAs, []core.Metric{})
			So(len(mts), ShouldBeGreaterThan, 0)
			So(len(mts[0].Config().Table()), ShouldBeGreaterThan, 0)
		})

		Convey("CollectMetrics provided a valid config", func() {
			cdn := cdata.NewNode()
			cdn.AddItem("someInt", ctypes.ConfigValueInt{Value: 1})
			cdn.AddItem("password", ctypes.ConfigValueStr{Value: "secure"})

			time.Sleep(500 * time.Millisecond)
			mts, err := c.CollectMetrics([]core.Metric{
				&plugin.MetricType{
					Namespace_: core.NewNamespace("foo", "bar"),
					Config_:    cdn,
				},
			})
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
			So(mts, ShouldHaveSameTypeAs, []core.Metric{})
			So(len(mts), ShouldBeGreaterThan, 0)
			So(mts[0].Config().Table(), ShouldNotBeEmpty)
			So(mts[0].Config().Table()["someInt"].Type(), ShouldResemble, "integer")

			Convey("Get and process the ConfigPolicy", func() {
				cp, err := c.GetConfigPolicy()
				So(err, ShouldBeNil)
				So(cp, ShouldNotBeNil)
				So(cp.Get([]string{"foo", "bar"}), ShouldNotBeNil)
				node := cp.Get([]string{"foo", "bar"})
				So(node, ShouldNotBeNil)
				cpn, cserrs := node.Process(mts[0].Config().Table())
				So(cpn, ShouldNotBeNil)
				So((*cpn)["somefloat"].Type(), ShouldResemble, "float")
				So((*cpn)["somefloat"].(ctypes.ConfigValueFloat).Value, ShouldResemble, 3.14)
				So(cserrs.Errors(), ShouldBeEmpty)
			})
		})

		Convey("CollectMetrics provided an invalid config", func() {
			cdn := cdata.NewNode()
			cdn.AddItem("someInt", ctypes.ConfigValueInt{Value: 1})

			time.Sleep(500 * time.Millisecond)
			mts, err := c.CollectMetrics([]core.Metric{
				&plugin.MetricType{
					Namespace_: core.NewNamespace("foo", "bar"),
					Config_:    cdn,
				},
			})
			So(err, ShouldBeNil)
			So(mts, ShouldNotBeNil)
			So(mts, ShouldHaveSameTypeAs, []core.Metric{})
			So(len(mts), ShouldBeGreaterThan, 0)
			So(mts[0].Config().Table(), ShouldNotBeEmpty)
			So(mts[0].Config().Table()["someInt"].Type(), ShouldResemble, "integer")

			Convey("Get and process the ConfigPolicy", func() {
				cp, err := c.GetConfigPolicy()
				So(err, ShouldBeNil)
				So(cp, ShouldNotBeNil)
				node := cp.Get([]string{"foo", "bar"})
				So(node, ShouldNotBeNil)
				So(err, ShouldBeNil)
				_, cserrs := node.Process(mts[0].Config().Table())
				//So(cpn, ShouldBeNil)
				So(cserrs.Errors(), ShouldNotBeEmpty)
				So(len(cserrs.Errors()), ShouldEqual, 1)
				So(cserrs.Errors()[0].Error(), ShouldContainSubstring, "password")
			})
		})
	})
}
