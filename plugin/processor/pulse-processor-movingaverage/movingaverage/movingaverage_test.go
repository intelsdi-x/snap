package movingaverage

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	"github.com/intelsdi-x/pulse/core/ctypes"
	. "github.com/smartystreets/goconvey/convey"
)

//Random number generator
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func TestMovingAverageProcessorMetrics(t *testing.T) {
	Convey("Moving Average Processor tests", t, func() {
		metrics := make([]plugin.PluginMetricType, 10)
		config := make(map[string]ctypes.ConfigValue)

		config["MovingAvgBufLength"] = ctypes.ConfigValueInt{Value: -1}

		Convey("Moving average for int data", func() {
			for i, _ := range metrics {
				time.Sleep(3)
				rand.Seed(time.Now().UTC().UnixNano())
				data := randInt(65, 90)
				metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", data)
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			movingAverageObj := NewMovingaverageProcessor()

			_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), config)

			var metrics_new []plugin.PluginMetricType

			//Decodes the content into pluginMetricType
			dec := gob.NewDecoder(bytes.NewBuffer(received_data))
			dec.Decode(&metrics_new)
			So(metrics, ShouldNotResemble, metrics_new)

		})

		Convey("Moving average for float32 data", func() {
			config["MovingAvgBufLength"] = ctypes.ConfigValueInt{Value: 40}
			for i, _ := range metrics {
				time.Sleep(3)
				rand.Seed(time.Now().UTC().UnixNano())
				data := randInt(65, 90)
				metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", float32(data))
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)

			movingAverageObj := NewMovingaverageProcessor()

			_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), config)

			var metrics_new []plugin.PluginMetricType

			//Decodes the content into pluginMetricType
			dec := gob.NewDecoder(bytes.NewBuffer(received_data))
			dec.Decode(&metrics_new)
			So(metrics, ShouldNotResemble, metrics_new)

		})
		Convey("Moving average for float64 data", func() {
			for i, _ := range metrics {
				time.Sleep(3)
				rand.Seed(time.Now().UTC().UnixNano())
				data := randInt(65, 90)
				metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", float64(data))
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)

			movingAverageObj := NewMovingaverageProcessor()

			_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), nil)

			var metrics_new []plugin.PluginMetricType

			//Decodes the content into pluginMetricType
			dec := gob.NewDecoder(bytes.NewBuffer(received_data))
			dec.Decode(&metrics_new)
			So(metrics, ShouldNotResemble, metrics_new)

		})

		Convey("Moving average for uint32 data", func() {
			for i, _ := range metrics {
				time.Sleep(3)
				rand.Seed(time.Now().UTC().UnixNano())
				data := randInt(65, 90)
				metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", uint32(data))
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			movingAverageObj := NewMovingaverageProcessor()

			_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), nil)

			var metrics_new []plugin.PluginMetricType

			//Decodes the content into pluginMetricType
			dec := gob.NewDecoder(bytes.NewBuffer(received_data))
			dec.Decode(&metrics_new)
			So(metrics, ShouldNotResemble, metrics_new)

		})

		Convey("Moving average for uint64 data", func() {
			for i, _ := range metrics {
				time.Sleep(3)
				rand.Seed(time.Now().UTC().UnixNano())
				data := randInt(65, 90)
				metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", uint64(data))
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)
			movingAverageObj := NewMovingaverageProcessor()

			_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), nil)

			var metrics_new []plugin.PluginMetricType

			//Decodes the content into pluginMetricType
			dec := gob.NewDecoder(bytes.NewBuffer(received_data))
			dec.Decode(&metrics_new)
			So(metrics, ShouldNotResemble, metrics_new)

		})

		Convey("Moving average for unknown data type", func() {
			for i, _ := range metrics {

				data := "I am an unknow data Type"
				metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, time.Now(), "", data)
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(metrics)

			movingAverageObj := NewMovingaverageProcessor()

			_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), nil)

			var metrics_new []plugin.PluginMetricType

			//Decodes the content into pluginMetricType
			dec := gob.NewDecoder(bytes.NewBuffer(received_data))
			dec.Decode(&metrics_new)
			So(metrics, ShouldNotResemble, metrics_new)

		})

	})
}
