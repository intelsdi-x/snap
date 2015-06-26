package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/scheduler"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PULSE_PATH        = os.Getenv("PULSE_PATH")
	DUMMY_PLUGIN_PATH = PULSE_PATH + "/plugin/collector/pulse-collector-dummy1"
)

func command() string {
	return "curl"
}

func readBody(r *http.Response) []byte {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()
	return b
}

func getAPIResponse(resp *http.Response) *APIResponse {
	r := new(APIResponse)
	rb := readBody(resp)
	err := json.Unmarshal(rb, r)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func TestPluginRestCalls(t *testing.T) {
	// Start a REST API to talk to
	log.SetLevel(log.WarnLevel)
	r := New()
	c := control.New()
	c.Start()
	s := scheduler.New()
	s.SetMetricManager(c)
	s.Start()
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	r.Start(":8181")
	time.Sleep(time.Millisecond * 100)

	Convey("REST API functional V1", t, func() {
		Convey("Load plugin - POST - /v1/plugins", func() {
			client := &http.Client{}
			file, err := os.Open(DUMMY_PLUGIN_PATH)
			if err != nil {
				log.Fatal(err)
			}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("pulse-plugins", filepath.Base(DUMMY_PLUGIN_PATH))
			if err != nil {
				log.Fatal(err)
			}
			_, err = io.Copy(part, file)
			if err != nil {
				log.Fatal(err)
			}

			err = writer.Close()
			if err != nil {
				log.Fatal(err)
			}

			req, err := http.NewRequest("POST", "http://localhost:8181/v1/plugins", body)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Add("Content-Type", writer.FormDataContentType())
			file.Close()
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			r := getAPIResponse(resp)
			plr := r.Body.(*rbody.PluginsLoaded)

			So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginsLoadedType)
			So(plr.ResponseBodyMessage(), ShouldEqual, "Plugins loaded: dummy1(collector v1)")
			So(len(plr.LoadedPlugins), ShouldEqual, 1)
			So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
			So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
			So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
			So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
			So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())

		})
		Convey("Plugin List - GET - /v1/plugins", func() {
			resp, err := http.Get("http://localhost:8181/v1/plugins")
			if err != nil {
				log.Fatal(err)
			}

			r := getAPIResponse(resp)
			plr := r.Body.(*rbody.PluginListReturned)

			So(plr.ResponseBodyType(), ShouldEqual, rbody.PluginListReturnedType)
			So(plr.ResponseBodyMessage(), ShouldEqual, "Plugin list retrieved")
			So(len(plr.LoadedPlugins), ShouldEqual, 1)
			So(plr.LoadedPlugins[0].Name, ShouldEqual, "dummy1")
			So(plr.LoadedPlugins[0].Version, ShouldEqual, 1)
			So(plr.LoadedPlugins[0].Status, ShouldEqual, "loaded")
			So(plr.LoadedPlugins[0].Type, ShouldEqual, "collector")
			So(plr.LoadedPlugins[0].LoadedTimestamp, ShouldBeLessThanOrEqualTo, time.Now().Unix())
		})

	})
}
