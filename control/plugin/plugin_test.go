package plugin

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"

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
		m := NewPluginMetricType([]string{"foo", "bar"})
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

func TestSessionState(t *testing.T) {
	Convey("SessionState", t, func() {
		now := time.Now()
		ss := &SessionState{
			LastPing: now,
			Arg:      &Arg{PingTimeoutDuration: 500 * time.Millisecond},
		}
		flag := true
		ss.logger = log.New(os.Stdout, ">>>", log.Ldate|log.Ltime)
		Convey("Ping", func() {

			ss.Ping(PingArgs{}, &flag)
			So(ss.LastPing.Nanosecond(), ShouldBeGreaterThan, now.Nanosecond())
		})
		Convey("Kill", func() {
			wtf := ss.Kill(KillArgs{Reason: "testing"}, &flag)
			So(wtf, ShouldBeNil)
		})
		Convey("GenerateResponse", func() {
			r := &Response{}
			ss.listenAddress = "1234"
			ss.token = "asdf"
			response := ss.generateResponse(r)
			So(response, ShouldHaveSameTypeAs, []byte{})
			json.Unmarshal(response, &r)
			So(r.ListenAddress, ShouldEqual, "1234")
			So(r.Token, ShouldEqual, "asdf")
		})
		Convey("InitSessionState", func() {
			var mockPluginArgs string = "{\"RunAsDaemon\": true, \"PingTimeoutDuration\": 2000000000}"
			sessionState, err, rc := NewSessionState(mockPluginArgs)
			So(sessionState.ListenAddress(), ShouldEqual, "")
			So(rc, ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(sessionState, ShouldNotBeNil)
			So(sessionState.PingTimeoutDuration, ShouldResemble, 2*time.Second)
		})
		Convey("InitSessionState with invalid args", func() {
			var mockPluginArgs string = ""
			_, err, _ := NewSessionState(mockPluginArgs)
			So(err, ShouldNotBeNil)
		})
		Convey("InitSessionState with a custom log path", func() {
			var mockPluginArgs string = "{\"RunAsDaemon\": false, \"PluginLogPath\": \"/var/tmp/pulse_plugin.log\"}"
			sessionState, err, rc := NewSessionState(mockPluginArgs)
			So(rc, ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(sessionState, ShouldNotBeNil)
		})
		Convey("heartbeatWatch timeout expired", func() {
			PingTimeoutLimit = 1
			ss.LastPing = now.Truncate(time.Minute)
			killChan := make(chan int)
			ss.heartbeatWatch(killChan)
			rc := <-killChan
			So(rc, ShouldEqual, 0)
		})
		Convey("heatbeatWatch reset", func() {
			PingTimeoutLimit = 2
			killChan := make(chan int)
			ss.heartbeatWatch(killChan)
			rc := <-killChan
			So(rc, ShouldEqual, 0)
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
	Convey("Start", t, func() {
		mockPluginMeta := NewPluginMeta("test", 1, CollectorPluginType)
		mockConfigPolicyTree := cpolicy.NewTree()
		var mockPluginArgs string = "{\"PluginLogPath\": \"/var/tmp/pulse_plugin.log\", \"PingTimeoutDuration\": 1000000000}"
		err, rc := Start(mockPluginMeta, new(MockPlugin), mockConfigPolicyTree, mockPluginArgs)
		So(err, ShouldBeNil)
		So(rc, ShouldEqual, 0)
	})
	Convey("Start with invalid args", t, func() {
		mockPluginMeta := NewPluginMeta("test", 1, CollectorPluginType)
		mockConfigPolicyTree := cpolicy.NewTree()
		var mockPluginArgs string = ""
		err, rc := Start(mockPluginMeta, new(MockPlugin), mockConfigPolicyTree, mockPluginArgs)
		So(err, ShouldNotBeNil)
		So(rc, ShouldNotEqual, 0)
	})
}
