/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Coporation

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
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/control/plugin/encoding"
	"github.com/intelsdi-x/pulse/control/plugin/encrypter"
	"github.com/intelsdi-x/pulse/core"
	"github.com/intelsdi-x/pulse/core/cdata"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	key, _    = rsa.GenerateKey(crand.Reader, 2048)
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
	var mts []plugin.PluginMetricType
	for _, i := range dargs.PluginMetricTypes {
		p := plugin.NewPluginMetricType(i.Namespace(), time.Now(), "", rand.Intn(100))
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
	pmts := []plugin.PluginMetricType{}
	pmts = append(pmts, plugin.PluginMetricType{
		Namespace_: []string{"foo", "bar"},
	})
	*reply, _ = m.e.Encode(plugin.GetMetricTypesReply{PluginMetricTypes: pmts})
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
	encr := encrypter.New(&key.PublicKey, key)
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
		c, err := NewCollectorHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second, &key.PublicKey, key)
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

			time.Sleep(500 * time.Millisecond)
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

			Convey("Get and process the ConfigPolicy", func() {
				cp, err := c.GetConfigPolicy()
				So(err, ShouldBeNil)
				So(cp, ShouldNotBeNil)
				So(cp.Get([]string{"foo", "bar"}), ShouldNotBeNil)
				node := cp.Get([]string{"foo", "bar"})
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

			time.Sleep(500 * time.Millisecond)
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

			Convey("Get and process the ConfigPolicy", func() {
				cp, err := c.GetConfigPolicy()
				So(err, ShouldBeNil)
				So(cp, ShouldNotBeNil)
				node := cp.Get([]string{"foo", "bar"})
				So(node, ShouldNotBeNil)
				So(err, ShouldBeNil)
				_, cperrs := node.Process(mts[0].Config().Table())
				//So(cpn, ShouldBeNil)
				So(cperrs.Errors(), ShouldNotBeEmpty)
				So(len(cperrs.Errors()), ShouldEqual, 1)
				So(cperrs.Errors()[0].Error(), ShouldContainSubstring, "password")
			})
		})
	})

	Convey("Processor Client", t, func() {
		session.c = false
		p, _ := NewProcessorHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second, &key.PublicKey, key)
		cl := p.(*httpJSONRPCClient)
		cl.encrypter.Key = symkey
		So(p, ShouldNotBeNil)

		Convey("GetConfigPolicy", func() {
			cp, err := p.GetConfigPolicy()
			So(err, ShouldBeNil)
			So(cp, ShouldNotBeNil)
			cp_ := cpolicy.New()
			cpn_ := cpolicy.NewPolicyNode()
			r1, err := cpolicy.NewIntegerRule("SomeRequiredInt", true, 1)
			r2, _ := cpolicy.NewStringRule("password", true)
			r3, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
			So(err, ShouldBeNil)
			cpn_.Add(r1, r2, r3)
			cp_.Add([]string{""}, cpn_)
			cpjson, _ := cp.MarshalJSON()
			cp_json, _ := cp_.MarshalJSON()
			So(string(cpjson), ShouldResemble, string(cp_json))
		})

		Convey("Process metrics", func() {
			pmt := plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", 1)
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

	Convey("Publisher Client", t, func() {
		session.c = false
		p, _ := NewPublisherHttpJSONRPCClient(fmt.Sprintf("http://%v", addr), 1*time.Second, &key.PublicKey, key)
		cl := p.(*httpJSONRPCClient)
		cl.encrypter.Key = symkey
		So(p, ShouldNotBeNil)

		Convey("GetConfigPolicy", func() {
			cp, err := p.GetConfigPolicy()
			So(err, ShouldBeNil)
			So(cp, ShouldNotBeNil)
			cp_ := cpolicy.New()
			cpn_ := cpolicy.NewPolicyNode()
			r1, err := cpolicy.NewIntegerRule("SomeRequiredInt", true, 1)
			r2, _ := cpolicy.NewStringRule("password", true)
			r3, _ := cpolicy.NewFloatRule("somefloat", false, 3.14)
			So(err, ShouldBeNil)
			cpn_.Add(r1, r2, r3)
			cp_.Add([]string{""}, cpn_)
			cpjson, _ := cp.MarshalJSON()
			cp_json, _ := cp_.MarshalJSON()
			So(string(cpjson), ShouldResemble, string(cp_json))
		})

		Convey("Publish metrics", func() {
			pmt := plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", 1)
			b, _ := json.Marshal([]plugin.PluginMetricType{*pmt})
			err := p.Publish(plugin.PulseJSONContentType, b, nil)
			So(err, ShouldBeNil)
		})

	})
}
