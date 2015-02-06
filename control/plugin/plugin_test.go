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
		now := time.Now().Unix()
		m := NewMetricType([]string{"foo", "bar"})
		Convey("New", func() {
			So(m, ShouldHaveSameTypeAs, &MetricType{})
		})
		Convey("Get Namespace", func() {
			So(m.Namespace, ShouldResemble, []string{"foo", "bar"})
		})
		Convey("Get LastAdvertisedTimestamp", func() {
			So(m.LastAdvertisedTimestamp, ShouldEqual, now)
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
		ss.Logger = log.New(os.Stdout, ">>>", log.Ldate|log.Ltime)
		Convey("Ping", func() {

			ss.Ping(PingArgs{}, &flag)
			So(ss.LastPing.Nanosecond(), ShouldBeGreaterThan, now.Nanosecond())
		})
		Convey("Kill", func() {
			wtf := ss.Kill(KillArgs{Reason: "testing"}, &flag)
			So(wtf, ShouldBeNil)
		})
		Convey("GenerateResponse", func() {
			r := Response{}
			ss.ListenAddress = "1234"
			ss.Token = "asdf"
			response := ss.GenerateResponse(r)
			So(response, ShouldHaveSameTypeAs, []byte{})
			json.Unmarshal(response, &r)
			So(r.ListenAddress, ShouldEqual, "1234")
			So(r.Token, ShouldEqual, "asdf")
		})
	})
}
