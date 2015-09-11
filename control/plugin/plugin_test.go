package plugin

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPluginType(t *testing.T) {
	Convey(".String()", t, func() {
		Convey("it returns the correct string representation", func() {
			p := PluginType(0)
			So(p.String(), ShouldEqual, "collector")
		})
	})
}

func TestMetricType(t *testing.T) {
	Convey("MetricType", t, func() {
		now := time.Now()
		m := &PluginMetricType{
			Namespace_:          []string{"foo", "bar"},
			LastAdvertisedTime_: now,
		}
		Convey("New", func() {
			So(m, ShouldHaveSameTypeAs, &PluginMetricType{})
		})
		Convey("Get Namespace", func() {
			So(m.Namespace(), ShouldResemble, []string{"foo", "bar"})
		})
		Convey("Get LastAdvertisedTimestamp", func() {
			So(m.LastAdvertisedTime().Unix(), ShouldBeGreaterThan, now.Unix()-2)
			So(m.LastAdvertisedTime().Unix(), ShouldBeLessThan, now.Unix()+2)
		})
	})
}

func TestArg(t *testing.T) {
	Convey("NewArg", t, func() {
		arg := NewArg(nil, "/tmp/pulse/plugin.log")
		So(arg, ShouldNotBeNil)
	})
}

func TestPlugin(t *testing.T) {
	a := []string{PulseAllContentType}
	b := []string{PulseGOBContentType}
	Convey("Start", t, func() {
		mockPluginMeta := NewPluginMeta("test", 1, CollectorPluginType, a, b)
		var mockPluginArgs string = "{\"PluginLogPath\": \"/var/tmp/pulse_plugin.log\"}"
		err, rc := Start(mockPluginMeta, new(MockPlugin), mockPluginArgs)
		So(err, ShouldBeNil)
		So(rc, ShouldEqual, 0)
	})
	Convey("Start with invalid args", t, func() {
		mockPluginMeta := NewPluginMeta("test", 1, CollectorPluginType, a, b)
		var mockPluginArgs string = ""
		err, rc := Start(mockPluginMeta, new(MockPlugin), mockPluginArgs)
		So(err, ShouldNotBeNil)
		So(rc, ShouldNotEqual, 0)
	})
	Convey("Bad accept type", t, func() {
		a := []string{"wat"}
		So(func() {
			NewPluginMeta("test", 1, CollectorPluginType, a, b)
		}, ShouldPanicWith, "Bad accept content type [test] for [1] [wat]")
	})
	Convey("Bad return type", t, func() {
		b := []string{"wat"}
		So(func() {
			NewPluginMeta("test", 1, CollectorPluginType, a, b)
		}, ShouldPanicWith, "Bad return content type [test] for [1] [wat]")
	})
}
