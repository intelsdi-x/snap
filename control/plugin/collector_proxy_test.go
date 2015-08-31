package plugin

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin/cpolicy"

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

func (p *mockPlugin) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	t := cpolicy.NewTree()
	cpn := cpolicy.NewPolicyNode()
	r1, _ := cpolicy.NewStringRule("username", false, "root")
	r2, _ := cpolicy.NewStringRule("password", true)
	cpn.Add(r1, r2)
	ns := []string{"one", "two", "potato"}
	t.Add(ns, cpn)
	t.Freeze()

	return *t, nil
}

type mockErrorPlugin struct {
}

func (p *mockErrorPlugin) GetMetricTypes() ([]PluginMetricType, error) {
	return nil, errors.New("Error in get Metric Type")
}

func (p *mockErrorPlugin) CollectMetrics(mockPluginMetricType []PluginMetricType) ([]PluginMetricType, error) {
	return nil, errors.New("Error in collect Metric")
}

func (p *mockErrorPlugin) GetConfigPolicyTree() (cpolicy.ConfigPolicyTree, error) {
	return cpolicy.ConfigPolicyTree{}, errors.New("Error in get config policy tree")
}

func TestCollectorProxy(t *testing.T) {
	Convey("Test collector plugin proxy for get metric types ", t, func() {

		logger := log.New(os.Stdout,
			"test: ",
			log.Ldate|log.Ltime|log.Lshortfile)
		mockPlugin := &mockPlugin{}

		mockSessionState := &MockSessionState{
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
			reply := &GetMetricTypesReply{
				PluginMetricTypes: nil,
			}
			c.GetMetricTypes(struct{}{}, reply)
			So(reply.PluginMetricTypes[0].Namespace(), ShouldResemble, []string{"foo", "bar"})

			Convey("Get error in Get Metric Type", func() {
				reply := &GetMetricTypesReply{
					PluginMetricTypes: nil,
				}
				mockErrorPlugin := &mockErrorPlugin{}
				errC := &collectorPluginProxy{
					Plugin:  mockErrorPlugin,
					Session: mockSessionState,
				}
				err := errC.GetMetricTypes(struct{}{}, reply)
				So(len(reply.PluginMetricTypes), ShouldResemble, 0)
				So(err.Error(), ShouldResemble, "GetMetricTypes call error : Error in get Metric Type")

			})

		})
		Convey("Collect Metric ", func() {
			args := CollectMetricsArgs{
				PluginMetricTypes: mockPluginMetricType,
			}
			reply := &CollectMetricsReply{
				PluginMetrics: nil,
			}
			c.CollectMetrics(args, reply)
			So(reply.PluginMetrics[0].Namespace(), ShouldResemble, []string{"foo", "bar"})

			Convey("Get error in Collect Metric ", func() {
				args := CollectMetricsArgs{
					PluginMetricTypes: mockPluginMetricType,
				}
				reply := &CollectMetricsReply{
					PluginMetrics: nil,
				}
				mockErrorPlugin := &mockErrorPlugin{}
				errC := &collectorPluginProxy{
					Plugin:  mockErrorPlugin,
					Session: mockSessionState,
				}
				err := errC.CollectMetrics(args, reply)
				So(len(reply.PluginMetrics), ShouldResemble, 0)
				So(err.Error(), ShouldResemble, "CollectMetrics call error : Error in collect Metric")

			})

		})
		Convey("Get Config Policy Tree", func() {
			replyPolicyTree := &GetConfigPolicyTreeReply{}

			c.GetConfigPolicyTree(struct{}{}, replyPolicyTree)

			So(replyPolicyTree.PolicyTree, ShouldNotBeNil)

			Convey("Get error in Config Policy Tree ", func() {
				mockErrorPlugin := &mockErrorPlugin{}
				errC := &collectorPluginProxy{
					Plugin:  mockErrorPlugin,
					Session: mockSessionState,
				}
				err := errC.GetConfigPolicyTree(struct{}{}, replyPolicyTree)
				So(err.Error(), ShouldResemble, "ConfigPolicyTree call error : Error in get config policy tree")

			})

		})

	})

}
