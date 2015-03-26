package control

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/intelsdilabs/gomit"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/pkg/logger"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	PluginPathDummy = path.Join(os.Getenv("PULSE_PATH"), "plugin", "collector", "pulse-collector-dummy")
	PluginPathMock  = path.Join(os.Getenv("PULSE_PATH"), "plugin", "collector", "pulse-collector-mock")
)

// ------------------ bunch of entities to just test CollectMetrics -----------

/*
cd pulse
make
cd control
export PULSE_PATH=`pwd`/../build
go test -run TestCollectMetrics
*/

// apEventWaiter is channel that waits for a availablePlugin started event
// a returns this instance of availablePlugin that was started
type apEventWaiter chan *availablePlugin

func (a apEventWaiter) HandleGomitEvent(event gomit.Event) {
	// fmt.Printf("EVENT = %#v\n", event.Body)
	switch ap := event.Body.(type) {
	case *availablePlugin:
		a <- ap
	}
}

var _ gomit.Handler = (*apEventWaiter)(nil) // check interface implementation

/*
steps:
2. load two plugins
*/
func TestCollectMetrics(t *testing.T) {

	// verbosity
	logger.SetLevel(logger.DebugLevel)
	// logger.SetLevel(logger.FatalLevel)
	logger.Output = os.Stdout

	// adjust HB timeouts for test
	//plugin.PingTimeoutLimit = 1
	//plugin.PingTimeoutDuration = time.Second * 1

	Convey("full scenario", t, func() {

		// --- Step1: new started pluginControl
		c := New()

		// attach my own handler for events to find a moment when availplugins are running
		waitForAP := make(apEventWaiter)
		c.eventManager.RegisterHandler("waitsforAP", waitForAP)

		// check handlers - we expect that handlers are prepared
		// make sure that runner waits for events as a handler
		So(c.pluginRunner, ShouldEqual, c.eventManager.Handlers["control.runner"])
		// pluginRunner can handle only MetricSubscriptionEvent

		// c.pluginRunner.(*runner).monitor.duration = time.Millisecond * 100
		c.Start()
		So(c.Started, ShouldBeTrue)
		// state of pluginRunner and his availplugins is empty
		So(len(*c.pluginRunner.AvailablePlugins().Collectors.table), ShouldEqual, 0)
		So(len(*c.pluginRunner.AvailablePlugins().Collectors.keys), ShouldEqual, 0)

		// --- Load plugin
		err := c.Load(PluginPathDummy) // causes LoadPluginEvent - but nobode cares
		So(err, ShouldBeNil)

		err = c.Load(PluginPathMock) // causes LoadPluginEvent - but nobode cares
		So(err, ShouldBeNil)

		// we expect two plugins loaded
		So(len(*c.pluginManager.LoadedPlugins().table), ShouldEqual, 2)

		// --- prepare metrics types we are intersted in
		// node metricconfiguration configuration

		mtGood, err := c.metricCatalog.Get([]string{"intel", "mock", "good"}, 0)
		So(err, ShouldBeNil)
		mtError, err := c.metricCatalog.Get([]string{"intel", "mock", "error"}, 0)
		So(err, ShouldBeNil)
		mtPanic, err := c.metricCatalog.Get([]string{"intel", "mock", "panic"}, 0)
		So(err, ShouldBeNil)
		mtTimeout, err := c.metricCatalog.Get([]string{"intel", "mock", "timeout"}, 0)
		So(err, ShouldBeNil)
		mtFoo, err := c.metricCatalog.Get([]string{"intel", "dummy", "foo"}, 0)
		So(err, ShouldBeNil)
		mtBar, err := c.metricCatalog.Get([]string{"intel", "dummy", "bar"}, 0)
		So(err, ShouldBeNil)

		// subscribe - actually is sends an event by gomit to runner to start available plugin
		cd := cdata.NewNode()

		for _, mt := range []core.MetricType{mtGood, mtError, mtPanic, mtTimeout, mtFoo, mtBar} {
			_, errs := c.SubscribeMetricType(mt, cd) // this events at the end starts a plugin
			So(errs, ShouldBeEmpty)
		}

		// expect two instances of availablePlugins
		for i := 0; i < 2; i++ {
			select {
			case <-waitForAP:
			case <-time.After(1 * time.Second):
				// give a second for each plugin to run and Fail if doen't expected event recevied
				t.Fail()
			}
		}
		fmt.Println("----------- two availablePlugins started !!!")

		// Call collect on router
		Convey("check when asked for all, get all", func() {
			metrics, errs := c.CollectMetrics([]core.MetricType{mtGood, mtFoo, mtBar}, cd, time.Now())
			So(errs, ShouldBeEmpty)
			So(metrics, ShouldNotBeEmpty)
			So(len(metrics), ShouldEqual, 2) // because actually dummy plugins always returns just foo - is it correct idk !
		})

		Convey("check when asked for broken get error", func() {
			metrics, errs := c.CollectMetrics([]core.MetricType{mtGood, mtFoo}, cd, time.Now())
			So(errs, ShouldNotBeEmpty)
			So(metrics, ShouldBeEmpty)
			fmt.Println(errs)
		})

		// because

		// fmt.Println(errs)

		// for x := 0; x < 5; x++ {
		// 	// fmt.Println("\n *  Calling Collect")
		// 	So(err, ShouldBeNil)
		// 	// fmt.Printf(" *  Collect Response: %+v\n", cr)
		// }
		// time.Sleep(time.Millisecond * 1000)
	})
}
