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

var mock_PluginMetricType []PluginMetricType = []PluginMetricType{
	*NewPluginMetricType([]string{"foo", "bar"}, 1),
	*NewPluginMetricType([]string{"foo", "baz"}, 2),
}

func (p *mockPlugin) GetMetricTypes() ([]PluginMetricType, error) {
	return mock_PluginMetricType, nil
}

func (p *mockPlugin) CollectMetrics(mock_PluginMetricType []PluginMetricType) ([]PluginMetricType, error) {
	return mock_PluginMetricType, nil
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

func (p *mockErrorPlugin) CollectMetrics(mock_PluginMetricType []PluginMetricType) ([]PluginMetricType, error) {
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
		mock_plugin := &mockPlugin{}

		mock_SessionState := &MockSessionState{
			listenPort:          "0",
			token:               "abcdef",
			logger:              logger,
			PingTimeoutDuration: time.Millisecond * 100,
			killChan:            make(chan int),
		}
		c := &collectorPluginProxy{
			Plugin:  mock_plugin,
			Session: mock_SessionState,
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
				mock_error_plugin := &mockErrorPlugin{}
				err_c := &collectorPluginProxy{
					Plugin:  mock_error_plugin,
					Session: mock_SessionState,
				}
				err := err_c.GetMetricTypes(struct{}{}, reply)
				So(len(reply.PluginMetricTypes), ShouldResemble, 0)
				So(err.Error(), ShouldResemble, "GetMetricTypes call error : Error in get Metric Type")

			})

		})
		Convey("Collect Metric ", func() {
			args := CollectMetricsArgs{
				PluginMetricTypes: mock_PluginMetricType,
			}
			reply := &CollectMetricsReply{
				PluginMetrics: nil,
			}
			c.CollectMetrics(args, reply)
			So(reply.PluginMetrics[0].Namespace(), ShouldResemble, []string{"foo", "bar"})

			Convey("Get error in Collect Metric ", func() {
				args := CollectMetricsArgs{
					PluginMetricTypes: mock_PluginMetricType,
				}
				reply := &CollectMetricsReply{
					PluginMetrics: nil,
				}
				mock_error_plugin := &mockErrorPlugin{}
				err_c := &collectorPluginProxy{
					Plugin:  mock_error_plugin,
					Session: mock_SessionState,
				}
				err := err_c.CollectMetrics(args, reply)
				So(len(reply.PluginMetrics), ShouldResemble, 0)
				So(err.Error(), ShouldResemble, "CollectMetrics call error : Error in collect Metric")

			})

		})
		Convey("Get Config Policy Tree", func() {
			reply_policy_tree := &GetConfigPolicyTreeReply{}

			c.GetConfigPolicyTree(struct{}{}, reply_policy_tree)

			So(reply_policy_tree.PolicyTree, ShouldNotBeNil)

			Convey("Get error in Config Policy Tree ", func() {
				mock_error_plugin := &mockErrorPlugin{}
				err_c := &collectorPluginProxy{
					Plugin:  mock_error_plugin,
					Session: mock_SessionState,
				}
				err := err_c.GetConfigPolicyTree(struct{}{}, reply_policy_tree)
				So(err.Error(), ShouldResemble, "ConfigPolicyTree call error : Error in get config policy tree")

			})

		})

	})

}
