// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

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

package strategy

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	. "github.com/intelsdi-x/snap/control/strategy/fixtures"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/core/serror"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPoolCreation(t *testing.T) {
	Convey("Given available collector type plugin", t, func() {
		plg := NewMockAvailablePlugin().WithVersion(3)
		Convey("When new plugin pool is being created with expected plugin type", func() {
			pool, err := NewPool(plg.String(), plg)
			Convey("Then new pool is created with no error", func() {
				So(pool, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("Then new pool is created with one plugin instance", func() {
				So(pool.Count(), ShouldEqual, 1)
			})
			Convey("Then new pool is created with plugin version", func() {
				So(pool.Version(), ShouldEqual, plg.Version())
			})
		})
	})

	Convey("Given available processor type plugin", t, func() {
		plg := NewMockAvailablePlugin().WithPluginType(plugin.ProcessorPluginType).WithVersion(1)
		Convey("When new plugin pool is being created with expected plugin type", func() {
			pool, err := NewPool(plg.String(), plg)
			Convey("Then new pool is created with no error", func() {
				So(pool, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("Then new pool is created with one plugin instance", func() {
				So(pool.Count(), ShouldEqual, 1)
			})
			Convey("Then new pool is created with plugin version", func() {
				So(pool.Version(), ShouldEqual, plg.Version())
			})
		})
	})

	Convey("Given available publisher type plugin", t, func() {
		plg := NewMockAvailablePlugin().WithPluginType(plugin.PublisherPluginType).WithVersion(2)
		Convey("When new plugin pool is being created with expected plugin type", func() {
			pool, err := NewPool(plg.String(), plg)
			Convey("Then new pool is created with no error", func() {
				So(pool, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("Then new pool is created with one plugin instance", func() {
				So(pool.Count(), ShouldEqual, 1)
			})
			Convey("Then new pool is created with plugin version", func() {
				So(pool.Version(), ShouldEqual, plg.Version())
			})
		})
	})

	Convey("Given available collector type plugin list", t, func() {
		plg := []AvailablePlugin{
			*NewMockAvailablePlugin().WithID(1),
			*NewMockAvailablePlugin().WithID(2),
		}
		Convey("When new plugin pool is being created with expected plugin type", func() {
			pool, err := NewPool(plg[0].String(), plg...)
			Convey("Then new pool is created with no error", func() {
				So(pool, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			Convey("Then new pool is created with one plugin instance", func() {
				So(pool.Count(), ShouldEqual, 2)
			})
			Convey("Then new pool is created with plugin version", func() {
				So(pool.Version(), ShouldEqual, plg[0].Version())
			})
		})
	})

	Convey("Given available collector type plugin", t, func() {
		plg := NewMockAvailablePlugin()
		Convey("When new plugin pool is being created with incorrect key", func() {
			badKey := plg.TypeName() + core.Separator + plg.Name()
			pool, err := NewPool(badKey, plg)
			Convey("Then pool is not created, error is not nil", func() {
				So(pool, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestPoolPluginMeta(t *testing.T) {
	Convey("Given available collector type plugin", t, func() {
		plg := NewMockAvailablePlugin().
			WithPluginType(plugin.CollectorPluginType).
			WithStrategy(plugin.DefaultRouting).
			WithExclusive(false).
			WithTTL(time.Second).
			WithConCount(2).
			WithID(1).
			WithVersion(2)
		Convey("When new plugin pool is being created with expected plugin type", func() {
			pool, _ := NewPool(plg.String(), plg)
			Convey("Then new pool has proper meta", func() {
				So(pool.Strategy().String(), ShouldEqual, plg.RoutingStrategy().String())
				So(pool.String(), ShouldEqual, plg.RoutingStrategy().String())
				So(pool.Eligible(), ShouldBeFalse)
				So(len(pool.Plugins()), ShouldEqual, 1)
				So(pool.Plugins()[1].String(), ShouldEqual, plg.String())
			})
		})
	})
}

func TestPoolEligibility(t *testing.T) {
	Convey("Given available collector type plugin", t, func() {
		plg := NewMockAvailablePlugin()
		Convey("When new plugin pool is being created with expected plugin type", func() {
			tcs := []struct {
				PlgType       plugin.PluginType
				Strategy      plugin.RoutingStrategyType
				Concurrency   int
				Subscriptions int
				Exclusiveness bool
				Expected      bool
			}{
				// type, strategy, concurrency count, number of subscriptions, exclusiveness, eligibility
				{plugin.CollectorPluginType, plugin.DefaultRouting, 1, 0, false, false},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 1, 1, false, false},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 2, 1, false, false},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 3, 1, false, false},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 1, 2, false, true},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 1, 3, false, true},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 3, 3, false, false},
				{plugin.CollectorPluginType, plugin.DefaultRouting, 1, 3, true, false},

				{plugin.CollectorPluginType, plugin.StickyRouting, 999, 0, false, false},
				{plugin.CollectorPluginType, plugin.StickyRouting, 999, 1, false, false},
				{plugin.CollectorPluginType, plugin.StickyRouting, 999, 2, false, true},
				{plugin.CollectorPluginType, plugin.StickyRouting, 999, 3, false, true},
				{plugin.CollectorPluginType, plugin.StickyRouting, 999, 4, false, true},
				{plugin.CollectorPluginType, plugin.StickyRouting, 999, 2, true, false},

				{plugin.CollectorPluginType, plugin.ConfigRouting, 1, 0, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 1, 1, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 1, 2, false, true},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 2, 1, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 2, 2, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 2, 3, false, true},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 2, 3, true, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 2, 4, false, true},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 3, 1, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 3, 2, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 3, 3, false, false},
				{plugin.CollectorPluginType, plugin.ConfigRouting, 3, 4, false, true},
			}
			Convey("Then new pool eligibility is defined", func() {
				for i, tc := range tcs {
					plg.WithPluginType(tc.PlgType).
						WithStrategy(tc.Strategy).
						WithExclusive(tc.Exclusiveness).
						WithConCount(tc.Concurrency).
						WithID(uint32(i))

					pool, _ := NewPool(plg.String(), plg)

					for j := 0; j < tc.Subscriptions; j++ {
						pool.Subscribe(strconv.Itoa(j))
					}

					Convey(fmt.Sprintf(
						"{strategy = %s, concurreny = %d, subscriptions = %d, exclusiveness = %v, count = %d}",
						tc.Strategy.String(),
						tc.Concurrency,
						tc.Subscriptions,
						tc.Exclusiveness,
						pool.Count(),
					),
						func() {
							So(pool.SubscriptionCount(), ShouldEqual, tc.Subscriptions)
							So(pool.Eligible(), ShouldEqual, tc.Expected)
						})
				}
			})
		})
	})
}

func TestPoolSelectAPDefaultRouter(t *testing.T) {
	Convey("For plugin defined with default strategy", t, func() {
		plugin := NewMockAvailablePlugin().WithStrategy(plugin.DefaultRouting)
		pool, _ := NewPool(plugin.String(), plugin)

		Convey("Then AvailablePlugin is selected", func() {
			ap, err := pool.SelectAP("TaskID", nil)
			So(ap, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func TestPoolSelectAPConfigRouter(t *testing.T) {
	Convey("Given task id and configuration", t, func() {
		cfg := map[string]ctypes.ConfigValue{"foo": ctypes.ConfigValueStr{"bar"}}
		otherCfg := map[string]ctypes.ConfigValue{"foo": ctypes.ConfigValueStr{"baz"}}

		Convey("When plugin is defined with config based strategy", func() {
			plugin := NewMockAvailablePlugin().WithStrategy(plugin.ConfigRouting)
			pool, _ := NewPool(plugin.String(), plugin)

			Convey("Then given routering is handled", func() {
				ap, err := pool.SelectAP("TaskID", cfg)
				So(ap, ShouldNotBeNil)
				So(err, ShouldBeNil)

				ap, err = pool.SelectAP("AnotherTaskID", cfg)
				So(ap, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(ap, ShouldEqual, plugin)

				ap, err = pool.SelectAP("YetAnotherTaskID", otherCfg)
				So(ap, ShouldBeNil)
				So(err, ShouldResemble, serror.New(ErrCouldNotSelect))
			})
		})

		Convey("When another plugin is defined with config based strategy", func() {
			plugin := NewMockAvailablePlugin().WithStrategy(plugin.ConfigRouting)
			pool, _ := NewPool(plugin.String(), plugin)

			Convey("With empty config, for some task, then routing is handled", func() {
				ap, err := pool.SelectAP("TaskID", map[string]ctypes.ConfigValue{})
				So(ap, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestPoolSelectAPStickyRouter(t *testing.T) {
	Convey("For plugin defined with sticky strategy", t, func() {
		plugin := NewMockAvailablePlugin().WithStrategy(plugin.StickyRouting)
		pool, _ := NewPool(plugin.String(), plugin)

		Convey("With empty config, for some task, routering is handled", func() {
			ap1, err := pool.SelectAP("TaskID", nil)
			So(ap1, ShouldNotBeNil)
			So(err, ShouldBeNil)

			cfg := map[string]ctypes.ConfigValue{"foo": ctypes.ConfigValueStr{"bar"}}
			ap2, err := pool.SelectAP("TaskID", cfg)
			So(ap2, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(ap2, ShouldEqual, ap1)

			ap3, err := pool.SelectAP("AnotherTaskID", nil)
			So(ap3, ShouldBeNil)
			So(err, ShouldResemble, serror.New(ErrCouldNotSelect))
		})
	})
}
