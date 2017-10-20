// +build medium

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

package control

import (
	"fmt"
	"net"
	"path"
	"testing"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/control_event"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/plugin/helper"

	"github.com/intelsdi-x/gomit"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestComparePlugins(t *testing.T) {
	Convey("Given new collector:plg:2 and old collector:plg:1", t, func() {
		plg2 := mockSubscribedPlugin{
			typeName: core.CollectorPluginType,
			name:     "plg",
			version:  2,
			config:   cdata.NewNode(),
		}

		plg1 := mockSubscribedPlugin{
			typeName: core.CollectorPluginType,
			name:     "plg",
			version:  1,
			config:   cdata.NewNode(),
		}
		Convey("When comparing new and old plugins", func() {
			adds, removes := comparePlugins([]core.SubscribedPlugin{plg2}, []core.SubscribedPlugin{plg1})

			Convey("Plugins to add and plugins to remove have proper elements", func() {
				So(len(adds), ShouldEqual, 1)
				So(len(removes), ShouldEqual, 1)
				So(subscribedPluginsContain(adds, plg2), ShouldBeTrue)
				So(subscribedPluginsContain(adds, plg1), ShouldBeFalse)
				So(subscribedPluginsContain(removes, plg2), ShouldBeFalse)
				So(subscribedPluginsContain(removes, plg1), ShouldBeTrue)
			})
		})
	})

	Convey("Given new collector:plg:2 and collector:plg:1 and old collector:plg:1", t, func() {
		plg1 := mockSubscribedPlugin{
			typeName: core.CollectorPluginType,
			name:     "plg",
			version:  2,
			config:   cdata.NewNode(),
		}

		plg2 := mockSubscribedPlugin{
			typeName: core.CollectorPluginType,
			name:     "plg",
			version:  1,
			config:   cdata.NewNode(),
		}
		Convey("When comparing new and old plugins", func() {
			adds, removes := comparePlugins([]core.SubscribedPlugin{plg1, plg2}, []core.SubscribedPlugin{plg1})

			Convey("Plugins to add and plugins to remove have proper elements", func() {
				So(len(adds), ShouldEqual, 1)
				So(len(removes), ShouldEqual, 0)
				So(subscribedPluginsContain(adds, plg2), ShouldBeTrue)
				So(subscribedPluginsContain(adds, plg1), ShouldBeFalse)
			})
		})
	})

	Convey("Given new collector:plg:1 and old collector:plg:2 and collector:plg:1", t, func() {
		plg1 := mockSubscribedPlugin{
			typeName: core.CollectorPluginType,
			name:     "plg",
			version:  2,
			config:   cdata.NewNode(),
		}

		plg2 := mockSubscribedPlugin{
			typeName: core.CollectorPluginType,
			name:     "plg",
			version:  1,
			config:   cdata.NewNode(),
		}
		Convey("When comparing new and old plugins", func() {
			adds, removes := comparePlugins([]core.SubscribedPlugin{plg1}, []core.SubscribedPlugin{plg1, plg2})

			Convey("Plugins to add and plugins to remove have proper elements", func() {
				So(len(adds), ShouldEqual, 0)
				So(len(removes), ShouldEqual, 1)
				So(subscribedPluginsContain(removes, plg2), ShouldBeTrue)
				So(subscribedPluginsContain(removes, plg1), ShouldBeFalse)
			})
		})
	})
}

func TestSubscriptionGroups_Process_GlobalPluginConfig(t *testing.T) {
	c := New(getTestSGConfig())
	Convey("Adds global plugin config for collectors", t, func() {
		c.Config.Plugins.Collector.All.AddItem("name", ctypes.ConfigValueStr{Value: "jane"})

		lpe := newLstnToPluginEvents()
		c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
		c.Start()

		Convey("Loading a mock collector plugin", func() {
			_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
			So(err, ShouldBeNil)
			<-lpe.load

			Convey("Subscription group created", func() {
				requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock", "foo")}
				subsPlugin := mockSubscribedPlugin{
					typeName: core.CollectorPluginType,
					name:     "mock",
					version:  1,
					config:   cdata.NewNode(),
				}

				sg := newSubscriptionGroups(c)
				So(sg, ShouldNotBeNil)
				sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{})
				<-lpe.sub
				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(len(group.plugins), ShouldEqual, 1)
				So(subscribedPluginsContain(group.plugins, subsPlugin), ShouldBeTrue)
				key := fmt.Sprintf("%s"+core.Separator+"%s"+core.Separator+"%d",
					subsPlugin.TypeName(),
					subsPlugin.Name(),
					subsPlugin.Version())
				So(group.metrics, ShouldContainKey, key)
				So(len(group.metrics[key].Metrics()), ShouldEqual, 1)
				So(group.metrics[key].Metrics()[0].Config().Table(), ShouldContainKey, "name")
				So(group.metrics[key].Metrics()[0].Config().Table()["name"], ShouldResemble, ctypes.ConfigValueStr{Value: "jane"})
			})
		})
	})
}

func TestSubscriptionGroups_ProcessStaticNegative(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
	c.Start()

	Convey("Loading a mock collector plugin", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with no wildcards", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock", "foo")}
			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			group, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(group, ShouldNotBeNil)
			So(len(group.plugins), ShouldEqual, 1)
			So(subscribedPluginsContain(group.plugins, subsPlugin), ShouldBeTrue)
			plgKey := key(group.plugins[0])
			So(group.metrics, ShouldContainKey, plgKey)
			metrics := group.metrics[plgKey].Metrics()
			So(len(metrics), ShouldEqual, 1)
			So(len(group.requestedMetrics), ShouldEqual, 1)
			So(metrics[0].Namespace().String(), ShouldEqual, group.requestedMetrics[0].Namespace().String())

			Convey("loading another mock", func() {
				anotherMock := mockSubscribedPlugin{
					typeName: core.CollectorPluginType,
					name:     "anothermock",
					version:  1,
					config:   cdata.NewNode(),
				}

				_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-anothermock1"))
				So(err, ShouldBeNil)
				<-lpe.load
				serrs := sg.Process()
				So(len(serrs), ShouldEqual, 0)

				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(len(group.plugins), ShouldEqual, 1)
				So(subscribedPluginsContain(group.plugins, subsPlugin), ShouldBeTrue)
				So(subscribedPluginsContain(group.plugins, anotherMock), ShouldBeFalse)
				plgKey := key(group.plugins[0])
				So(group.metrics, ShouldContainKey, plgKey)
				metrics := group.metrics[plgKey].Metrics()
				So(len(metrics), ShouldEqual, 1)
				So(len(group.requestedMetrics), ShouldEqual, 1)
				So(metrics[0].Namespace().String(), ShouldEqual, group.requestedMetrics[0].Namespace().String())
			})
		})
	})
}

func TestSubscriptionGroups_ProcessStaticPositive(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
	c.Start()

	Convey("Loading a mock collector plugin", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with no wildcards", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock", "foo")}
			mock1 := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			group, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(group, ShouldNotBeNil)
			So(len(group.plugins), ShouldEqual, 1)
			So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
			plgKey := key(group.plugins[0])
			So(group.metrics, ShouldContainKey, plgKey)
			metrics := group.metrics[plgKey].Metrics()
			So(len(metrics), ShouldEqual, 1)
			So(len(group.requestedMetrics), ShouldEqual, 1)
			So(metrics[0].Namespace().String(), ShouldEqual, group.requestedMetrics[0].Namespace().String())

			Convey("loading another mock", func() {
				mock2 := mockSubscribedPlugin{
					typeName: core.CollectorPluginType,
					name:     "mock",
					version:  2,
					config:   cdata.NewNode(),
				}
				_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock2"))
				So(err, ShouldBeNil)
				<-lpe.load
				serrs := sg.Process()
				So(len(serrs), ShouldEqual, 0)

				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(len(group.plugins), ShouldEqual, 1)
				So(subscribedPluginsContain(group.plugins, mock1), ShouldBeFalse)
				So(subscribedPluginsContain(group.plugins, mock2), ShouldBeTrue)
				plgKey := key(group.plugins[0])
				So(group.metrics, ShouldContainKey, plgKey)
				metrics := group.metrics[plgKey].Metrics()
				So(len(metrics), ShouldEqual, 1)
				So(len(group.requestedMetrics), ShouldEqual, 1)
				So(metrics[0].Namespace().String(), ShouldEqual, group.requestedMetrics[0].Namespace().String())
			})
		})
	})
}

func TestSubscriptionGroups_ProcessDynamicPositive(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
	c.Start()

	Convey("Loading a mock collector plugin", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)

		<-lpe.load

		Convey("ValidateDeps", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel").AddDynamicElement("wild", "wild description")}
			mock1 := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}
			cnode := cdata.NewNode()
			cnode.AddItem("password", ctypes.ConfigValueStr{Value: "secret"})
			ctree := cdata.NewTree()
			ctree.Add([]string{"intel", "mock"}, cnode)
			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			errs := sg.ValidateDeps([]core.RequestedMetric{requested}, []core.SubscribedPlugin{mock1}, ctree)
			So(errs, ShouldBeNil)
			Convey("Subscription group created for requested metric with wildcards", func() {
				sg.Add("task-id", []core.RequestedMetric{requested}, ctree, []core.SubscribedPlugin{})
				<-lpe.sub
				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(len(group.plugins), ShouldEqual, 1)
				So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
				plgKey := key(group.plugins[0])
				So(group.metrics, ShouldContainKey, plgKey)
				metrics := group.metrics[plgKey].Metrics()
				So(len(metrics), ShouldBeGreaterThan, 1)
				So(len(group.requestedMetrics), ShouldEqual, 1)
				mts1 := len(metrics)

				Convey("loading another mock", func() {
					anotherMock1 := mockSubscribedPlugin{
						typeName: core.CollectorPluginType,
						name:     "anothermock",
						version:  1,
						config:   cdata.NewNode(),
					}
					_, err := loadPlg(c, path.Join(helper.PluginFilePath("snap-plugin-collector-anothermock1")))
					So(err, ShouldBeNil)
					<-lpe.load
					serrs := sg.Process()
					So(len(serrs), ShouldEqual, 0)

					So(len(sg.subscriptionMap), ShouldEqual, 1)
					group, ok := sg.subscriptionMap["task-id"]
					So(ok, ShouldBeTrue)
					So(group, ShouldNotBeNil)
					So(len(group.plugins), ShouldEqual, 2)
					So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
					So(subscribedPluginsContain(group.plugins, anotherMock1), ShouldBeTrue)
					plgKey1 := key(group.plugins[0])
					plgKey2 := key(group.plugins[1])
					So(group.metrics, ShouldContainKey, plgKey1)
					So(group.metrics, ShouldContainKey, plgKey2)
					metricsPlg1 := group.metrics[plgKey1].Metrics()
					metricsPlg2 := group.metrics[plgKey2].Metrics()
					So(len(metricsPlg1), ShouldBeGreaterThan, 1)
					So(len(metricsPlg2), ShouldBeGreaterThan, 1)
					So(len(group.requestedMetrics), ShouldEqual, 1)
					mts2 := len(metricsPlg1) + len(metricsPlg2)
					So(mts2, ShouldBeGreaterThan, mts1)
				})
			})
		})
	})
}

func TestSubscriptionGroups_ProcessDynamicNegative(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
	c.Start()

	Convey("Loading a mock collector plugin", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with wildcards", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock").AddDynamicElement("wild", "wild description")}
			mock1 := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			group, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(group, ShouldNotBeNil)
			So(len(group.plugins), ShouldEqual, 1)
			So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
			plgKey := key(group.plugins[0])
			So(group.metrics, ShouldContainKey, plgKey)
			metrics := group.metrics[plgKey].Metrics()
			So(len(metrics), ShouldBeGreaterThan, 1)
			So(len(group.requestedMetrics), ShouldEqual, 1)
			mts1 := len(metrics)

			Convey("loading another mock", func() {
				anotherMock1 := mockSubscribedPlugin{
					typeName: core.CollectorPluginType,
					name:     "anothermock",
					version:  1,
					config:   cdata.NewNode(),
				}
				_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-anothermock1"))
				So(err, ShouldBeNil)
				<-lpe.load
				serrs := sg.Process()
				So(len(serrs), ShouldEqual, 0)

				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(len(group.plugins), ShouldEqual, 1)
				So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
				So(subscribedPluginsContain(group.plugins, anotherMock1), ShouldBeFalse)
				plgKey := key(group.plugins[0])
				So(group.metrics, ShouldContainKey, plgKey)
				metrics := group.metrics[plgKey].Metrics()
				So(len(metrics), ShouldBeGreaterThan, 1)
				So(len(group.requestedMetrics), ShouldEqual, 1)
				mts2 := len(metrics)
				So(mts1, ShouldEqual, mts2)
			})
		})
	})
}

func TestSubscriptionGroups_ProcessSpecifiedDynamicPositive(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
	c.Start()

	Convey("Loading a mock collector plugin", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)

		<-lpe.load
		Convey("ValidateDeps", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel").AddDynamicElement("wild", "wild description").AddDynamicElement("host", "name of the host").AddStaticElement("baz")}
			// specified dynamic element
			requested.Namespace()[2].Value = "host0"
			mock1 := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			serrs := sg.ValidateDeps([]core.RequestedMetric{requested}, []core.SubscribedPlugin{mock1}, cdata.NewTree())
			So(serrs, ShouldBeNil)
			Convey("Subscription group created for requested metric with specified instance of dynamic element and with wildcards", func() {
				sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{})
				<-lpe.sub
				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
				So(len(group.plugins), ShouldEqual, 1)
				plgKey := key(group.plugins[0])
				So(group.metrics, ShouldContainKey, plgKey)
				metrics := group.metrics[plgKey].Metrics()
				// expected 1 subscribed metric: `/intel/mock/host0/baz`
				So(len(metrics), ShouldEqual, 1)
				So(len(group.requestedMetrics), ShouldEqual, 1)
				mts1 := len(metrics)

				Convey("loading another mock", func() {
					anotherMock1 := mockSubscribedPlugin{
						typeName: core.CollectorPluginType,
						name:     "anothermock",
						version:  1,
						config:   cdata.NewNode(),
					}
					_, err := loadPlg(c, path.Join(helper.PluginFilePath("snap-plugin-collector-anothermock1")))
					So(err, ShouldBeNil)
					<-lpe.load
					serrs := sg.Process()
					So(len(serrs), ShouldEqual, 0)

					So(len(sg.subscriptionMap), ShouldEqual, 1)
					group, ok := sg.subscriptionMap["task-id"]
					So(ok, ShouldBeTrue)
					So(group, ShouldNotBeNil)
					So(len(group.plugins), ShouldEqual, 2)
					So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
					So(subscribedPluginsContain(group.plugins, anotherMock1), ShouldBeTrue)
					plgKey1 := key(group.plugins[0])
					plgKey2 := key(group.plugins[1])
					So(group.metrics, ShouldContainKey, plgKey1)
					So(group.metrics, ShouldContainKey, plgKey2)
					metricsPlg1 := group.metrics[plgKey1].Metrics()
					metricsPlg2 := group.metrics[plgKey2].Metrics()
					// expected 1 subscribed metric per each plugin:
					// `/intel/mock/host0/baz` and `/intel/anothermock/host0/baz`
					So(len(metricsPlg1), ShouldEqual, 1)
					So(len(metricsPlg2), ShouldEqual, 1)

					mts2 := len(metricsPlg1) + len(metricsPlg2)
					So(mts2, ShouldBeGreaterThan, mts1)
				})
			})
		})
	})
}

func TestSubscriptionGroups_ProcessSpecifiedDynamicNegative(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_Process", lpe)
	c.Start()

	Convey("Loading a mock collector plugin", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with specified instance of dynamic element", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz")}
			// specified dynamic element
			requested.Namespace()[2].Value = "host0"
			mock1 := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			group, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(group, ShouldNotBeNil)
			So(len(group.plugins), ShouldEqual, 1)
			So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
			plgKey := key(group.plugins[0])
			So(group.metrics, ShouldContainKey, plgKey)
			metrics := group.metrics[plgKey].Metrics()
			// expected 1 subscribed metrics:`/intel/mock/host0/baz`
			So(len(metrics), ShouldEqual, 1)
			So(len(group.requestedMetrics), ShouldEqual, 1)
			So(metrics[0].Namespace().String(), ShouldEqual, group.requestedMetrics[0].Namespace().String())
			mts1 := len(metrics)

			Convey("loading another mock", func() {
				anotherMock1 := mockSubscribedPlugin{
					typeName: core.CollectorPluginType,
					name:     "anothermock",
					version:  1,
					config:   cdata.NewNode(),
				}
				_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-anothermock1"))
				So(err, ShouldBeNil)
				<-lpe.load
				serrs := sg.Process()
				So(len(serrs), ShouldEqual, 0)

				So(len(sg.subscriptionMap), ShouldEqual, 1)
				group, ok := sg.subscriptionMap["task-id"]
				So(ok, ShouldBeTrue)
				So(group, ShouldNotBeNil)
				So(len(group.plugins), ShouldEqual, 1)
				So(subscribedPluginsContain(group.plugins, mock1), ShouldBeTrue)
				So(subscribedPluginsContain(group.plugins, anotherMock1), ShouldBeFalse)
				plgKey := key(group.plugins[0])
				So(group.metrics, ShouldContainKey, plgKey)
				metrics := group.metrics[plgKey].Metrics()
				// expected 1 subscribed metrics:`/intel/mock/host0/baz`
				So(len(metrics), ShouldEqual, 1)
				So(len(group.requestedMetrics), ShouldEqual, 1)
				So(metrics[0].Namespace().String(), ShouldEqual, group.requestedMetrics[0].Namespace().String())
				mts2 := len(metrics)
				So(mts1, ShouldEqual, mts2)
			})
		})
	})
}

func TestSubscriptionGroups_AddRemoveStatic(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_AddRemove", lpe)
	c.Start()

	Convey("Loading a mock collector plugn", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with no wildcards", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock", "foo")}
			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{subsPlugin})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			val, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(val, ShouldNotBeNil)

			serrs := sg.Remove("task-id")
			<-lpe.unsub
			So(len(serrs), ShouldEqual, 0)
			So(len(sg.subscriptionMap), ShouldEqual, 0)
		})
	})
}

func TestSubscriptionGroups_AddRemoveDynamic(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_AddRemove", lpe)
	c.Start()

	Convey("Loading a mock collector plugn", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with wildcards", func() {
			requested := mockRequestedMetric{
				namespace: core.NewNamespace("intel", "mock").AddDynamicElement("wild", "wild description"),
				version:   -1,
			}
			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{subsPlugin})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			val, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(val, ShouldNotBeNil)

			serrs := sg.Remove("task-id")
			<-lpe.unsub
			So(len(serrs), ShouldEqual, 0)
			So(len(sg.subscriptionMap), ShouldEqual, 0)
		})
	})
}

func TestSubscriptionGroups_AddRemoveSpecifiedDynamic(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_AddRemove", lpe)
	c.Start()

	Convey("Loading a mock collector plugn", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with specified instance of dynamic element", func() {
			requested := mockRequestedMetric{
				namespace: core.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz"),
				version:   -1,
			}
			// specified dynamic element
			requested.Namespace()[2].Value = "host0"

			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{subsPlugin})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			val, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(val, ShouldNotBeNil)

			serrs := sg.Remove("task-id")
			<-lpe.unsub
			So(len(serrs), ShouldEqual, 0)
			So(len(sg.subscriptionMap), ShouldEqual, 0)
		})
	})
}

func TestSubscriptionGroups_GetStatic(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_AddRemove", lpe)
	c.Start()

	Convey("Loading a mock collector plugn", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with no wildcards", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock", "foo")}
			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}
			subsPluginKey := key(subsPlugin)

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{subsPlugin})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			val, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(val, ShouldNotBeNil)

			pluginToMetricMap, serrs, err := sg.Get("task-id")
			So(len(serrs), ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(len(pluginToMetricMap), ShouldEqual, 1)
			So(pluginToMetricMap, ShouldContainKey, subsPluginKey)
			metrics := pluginToMetricMap[subsPluginKey].Metrics()
			So(len(metrics), ShouldEqual, 1)

			pluginToMetricMap, serrs, err = sg.Get("task-fake-id")
			So(len(serrs), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrSubscriptionGroupDoesNotExist)
			So(pluginToMetricMap, ShouldBeEmpty)
		})
	})
}

func TestSubscriptionGroups_GetDynamic(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_AddRemove", lpe)
	c.Start()

	Convey("Loading a mock collector plugn", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with no wildcards", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel").AddDynamicElement("wild", "wild description")}
			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}
			subsPluginKey := key(subsPlugin)

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{subsPlugin})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			val, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(val, ShouldNotBeNil)

			pluginToMetricMap, serrs, err := sg.Get("task-id")
			So(len(serrs), ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(len(pluginToMetricMap), ShouldEqual, 1)
			So(pluginToMetricMap, ShouldContainKey, subsPluginKey)
			metrics := pluginToMetricMap[subsPluginKey].Metrics()
			So(len(metrics), ShouldBeGreaterThan, 1)

			pluginToMetricMap, serrs, err = sg.Get("task-fake-id")
			So(len(serrs), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrSubscriptionGroupDoesNotExist)
			So(pluginToMetricMap, ShouldBeEmpty)
		})
	})
}

func TestSubscriptionGroups_GetSpecifiedDynamic(t *testing.T) {
	c := New(getTestSGConfig())

	lpe := newLstnToPluginEvents()
	c.eventManager.RegisterHandler("TestSubscriptionGroups_AddRemove", lpe)
	c.Start()

	Convey("Loading a mock collector plugn", t, func() {
		_, err := loadPlg(c, helper.PluginFilePath("snap-plugin-collector-mock1"))
		So(err, ShouldBeNil)
		<-lpe.load

		Convey("Subscription group created for requested metric with specified instance of dynamic element", func() {
			requested := mockRequestedMetric{namespace: core.NewNamespace("intel", "mock").AddDynamicElement("host", "name of the host").AddStaticElement("baz")}
			// specified dynamic element
			requested.Namespace()[2].Value = "host0"
			subsPlugin := mockSubscribedPlugin{
				typeName: core.CollectorPluginType,
				name:     "mock",
				version:  1,
				config:   cdata.NewNode(),
			}
			subsPluginKey := key(subsPlugin)

			sg := newSubscriptionGroups(c)
			So(sg, ShouldNotBeNil)
			sg.Add("task-id", []core.RequestedMetric{requested}, cdata.NewTree(), []core.SubscribedPlugin{subsPlugin})
			<-lpe.sub
			So(len(sg.subscriptionMap), ShouldEqual, 1)
			val, ok := sg.subscriptionMap["task-id"]
			So(ok, ShouldBeTrue)
			So(val, ShouldNotBeNil)

			pluginToMetricMap, serrs, err := sg.Get("task-id")
			So(len(serrs), ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(len(pluginToMetricMap), ShouldEqual, 1)
			So(pluginToMetricMap, ShouldContainKey, subsPluginKey)
			metrics := pluginToMetricMap[subsPluginKey].Metrics()
			So(len(metrics), ShouldEqual, 1)

			pluginToMetricMap, serrs, err = sg.Get("task-fake-id")
			So(len(serrs), ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, ErrSubscriptionGroupDoesNotExist)
			So(pluginToMetricMap, ShouldBeEmpty)
		})
	})
}

type lstnToPluginEvents struct {
	load    chan struct{}
	sub     chan struct{}
	unsub   chan struct{}
	started chan struct{}
}

func newLstnToPluginEvents() *lstnToPluginEvents {
	return &lstnToPluginEvents{
		load:    make(chan struct{}),
		unsub:   make(chan struct{}),
		sub:     make(chan struct{}),
		started: make(chan struct{}),
	}
}

func (l *lstnToPluginEvents) HandleGomitEvent(e gomit.Event) {
	switch v := e.Body.(type) {
	case *control_event.PluginSubscriptionEvent:
		l.sub <- struct{}{}
	case *control_event.PluginUnsubscriptionEvent:
		l.unsub <- struct{}{}
	case *control_event.LoadPluginEvent:
		l.load <- struct{}{}
	case *control_event.StartPluginEvent:
		l.started <- struct{}{}
	default:
		controlLogger.WithFields(log.Fields{
			"event:": v.Namespace(),
			"_block": "HandleGomit",
		}).Info("Unhandled Event")
	}
}

func getTestSGConfig() *Config {
	config := GetDefaultConfig()
	config.ListenPort = getTestSGPort()
	return config
}

func getTestSGPort() int {
	count := 0
	for count < 1000 {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err == nil {
			p := ln.Addr().(*net.TCPAddr).Port
			ln.Close()
			return p
		}
		count++
	}
	panic("Could not find an available port")
}

func loadPlg(c *pluginControl, paths ...string) (core.CatalogedPlugin, serror.SnapError) {
	// This is a Travis optimized loading of plugins. From time to time, tests will error in Travis
	// due to a timeout when waiting for a response from a plugin. We are going to attempt loading a plugin
	// 3 times before letting the error through. Hopefully this cuts down on the number of Travis failures
	var e serror.SnapError
	var p core.CatalogedPlugin
	rp, err := core.NewRequestedPlugin(paths[0], GetDefaultConfig().TempDirPath, nil)
	if err != nil {
		return nil, serror.New(err)
	}
	if len(paths) > 1 {
		rp.SetSignature([]byte{00, 00, 00})
	}
	for i := 0; i < 3; i++ {
		p, e = c.Load(rp)
		if e == nil {
			break
		}
		if e != nil && i == 2 {
			return nil, e

		}
	}
	return p, nil
}

type mockSubscribedPlugin struct {
	typeName core.PluginType
	name     string
	version  int
	config   *cdata.ConfigDataNode
}

func (msp mockSubscribedPlugin) TypeName() string {
	return msp.typeName.String()
}

func (msp mockSubscribedPlugin) Name() string {
	return msp.name
}

func (msp mockSubscribedPlugin) Version() int {
	return msp.version
}

func (msp mockSubscribedPlugin) Config() *cdata.ConfigDataNode {
	return msp.config
}

type mockRequestedMetric struct {
	namespace core.Namespace
	version   int
}

func (mrm mockRequestedMetric) Namespace() core.Namespace {
	return mrm.namespace
}

func (mrm mockRequestedMetric) Version() int {
	return mrm.version
}

func subscribedPluginsContain(list []core.SubscribedPlugin, lookup core.SubscribedPlugin) bool {
	for _, plugin := range list {
		if plugin.TypeName() == lookup.TypeName() && plugin.Name() == lookup.Name() && plugin.Version() == lookup.Version() {
			return true
		}
	}
	return false
}
