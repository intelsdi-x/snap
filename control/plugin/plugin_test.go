package plugin

import (
	"encoding/json"
	"log"
	"os"
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
		m := NewMetricType([]string{"foo", "bar"}, now)
		Convey("New", func() {
			So(m, ShouldHaveSameTypeAs, &MetricType{})
		})
		Convey("Get Namespace", func() {
			So(m.Namespace(), ShouldResemble, []string{"foo", "bar"})
		})
		Convey("Get LastAdvertisedTimestamp", func() {
			So(m.LastAdvertisedTime().Unix(), ShouldEqual, now.Unix())
		})
	})
}

func TestSessionState(t *testing.T) {
	Convey("SessionState", t, func() {
		now := time.Now()
		ss := &SessionState{
			LastPing: now,
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
			var mockPluginArgs string = "{\"RunAsDaemon\": false}"
			sessionState, err, rc := NewSessionState(mockPluginArgs)
			So(sessionState.ListenAddress(), ShouldEqual, "")
			So(rc, ShouldEqual, 0)
			So(err, ShouldBeNil)
			So(sessionState, ShouldNotBeNil)
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
			PingTimeoutDuration = time.Millisecond * 100
			PingTimeoutLimit = 1
			ss.LastPing = now.Truncate(time.Minute)
			killChan := make(chan int)
			ss.heartbeatWatch(killChan)
			rc := <-killChan
			So(rc, ShouldEqual, 0)
		})
		Convey("heatbeatWatch reset", func() {
			PingTimeoutDuration = time.Millisecond * 500
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
		mockConfigPolicy := new(ConfigPolicy)
		var mockPluginArgs string = "{\"PluginLogPath\": \"/var/tmp/pulse_plugin.log\"}"
		err, rc := Start(mockPluginMeta, new(MockPlugin), mockConfigPolicy, mockPluginArgs)
		So(err, ShouldBeNil)
		So(rc, ShouldEqual, 0)
	})
	Convey("Start with invalid args", t, func() {
		mockPluginMeta := NewPluginMeta("test", 1, CollectorPluginType)
		mockConfigPolicy := new(ConfigPolicy)
		var mockPluginArgs string = ""
		err, rc := Start(mockPluginMeta, new(MockPlugin), mockConfigPolicy, mockPluginArgs)
		So(err, ShouldNotBeNil)
		So(rc, ShouldNotEqual, 0)
	})
}
