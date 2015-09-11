package plugin

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"
	"github.com/intelsdi-x/pulse/control/plugin/encoding"

	. "github.com/smartystreets/goconvey/convey"
)

type mockPlugin struct {
}

var mockPluginMetricType []PluginMetricType = []PluginMetricType{
	*NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", 1),
	*NewPluginMetricType([]string{"foo", "baz"}, time.Now(), "", 2),
}

func (p *mockPlugin) GetMetricTypes() ([]PluginMetricType, error) {
	return mockPluginMetricType, nil
}

func (p *mockPlugin) CollectMetrics(mockPluginMetricType []PluginMetricType) ([]PluginMetricType, error) {
	return mockPluginMetricType, nil
}

func (p *mockPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	cpn := cpolicy.NewPolicyNode()
	r1, _ := cpolicy.NewStringRule("username", false, "root")
	r2, _ := cpolicy.NewStringRule("password", true)
	cpn.Add(r1, r2)
	ns := []string{"one", "two", "potato"}
	cp.Add(ns, cpn)
	cp.Freeze()

	return cp, nil
}

type mockErrorPlugin struct {
}

func (p *mockErrorPlugin) GetMetricTypes() ([]PluginMetricType, error) {
	return nil, errors.New("Error in get Metric Type")
}

func (p *mockErrorPlugin) CollectMetrics(mockPluginMetricType []PluginMetricType) ([]PluginMetricType, error) {
	return nil, errors.New("Error in collect Metric")
}

func (p *mockErrorPlugin) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	return &cpolicy.ConfigPolicy{}, errors.New("Error in get config policy")
}

func TestCollectorProxy(t *testing.T) {
	Convey("Test collector plugin proxy for get metric types ", t, func() {

		logger := log.New(os.Stdout,
			"test: ",
			log.Ldate|log.Ltime|log.Lshortfile)
		mockPlugin := &mockPlugin{}

		mockSessionState := &MockSessionState{
			Encoder:             encoding.NewGobEncoder(),
			listenPort:          "0",
			token:               "abcdef",
			logger:              logger,
			PingTimeoutDuration: time.Millisecond * 100,
			killChan:            make(chan int),
		}
		c := &collectorPluginProxy{
			Plugin:  mockPlugin,
			Session: mockSessionState,
		}
		Convey("Get Metric Types", func() {
			var reply []byte
			c.GetMetricTypes([]byte{}, &reply)
			var mtr GetMetricTypesReply
			err := c.Session.Decode(reply, &mtr)
			So(err, ShouldBeNil)
			So(mtr.PluginMetricTypes[0].Namespace(), ShouldResemble, []string{"foo", "bar"})

		})
		Convey("Get error in Get Metric Type", func() {
			mockErrorPlugin := &mockErrorPlugin{}
			errC := &collectorPluginProxy{
				Plugin:  mockErrorPlugin,
				Session: mockSessionState,
			}
			var reply []byte
			err := errC.GetMetricTypes([]byte{}, &reply)
			So(err.Error(), ShouldResemble, "GetMetricTypes call error : Error in get Metric Type")
		})
		Convey("Collect Metric ", func() {
			args := CollectMetricsArgs{
				PluginMetricTypes: mockPluginMetricType,
			}
			out, err := c.Session.Encode(args)
			So(err, ShouldBeNil)
			var reply []byte
			c.CollectMetrics(out, &reply)
			var mtr CollectMetricsReply
			err = c.Session.Decode(reply, &mtr)
			So(mtr.PluginMetrics[0].Namespace(), ShouldResemble, []string{"foo", "bar"})

			Convey("Get error in Collect Metric ", func() {
				args := CollectMetricsArgs{
					PluginMetricTypes: mockPluginMetricType,
				}
				mockErrorPlugin := &mockErrorPlugin{}
				errC := &collectorPluginProxy{
					Plugin:  mockErrorPlugin,
					Session: mockSessionState,
				}
				out, err := errC.Session.Encode(args)
				So(err, ShouldBeNil)
				var reply []byte
				err = errC.CollectMetrics(out, &reply)
				So(err, ShouldNotBeNil)
			})

		})

	})

}
