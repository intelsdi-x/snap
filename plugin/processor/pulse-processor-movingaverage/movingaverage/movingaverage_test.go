package movingaverage

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/intelsdi-x/pulse/control/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

//Random number generator
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func TestMovingAverageProcessorMetrics(t *testing.T) {
	Convey("Moving Average Processor tests", t, func() {
		metrics := make([]plugin.PluginMetricType, 10)

		for i, _ := range metrics {
			time.Sleep(3)
			rand.Seed(time.Now().UTC().UnixNano())
			data := randInt(65, 90)
			metrics[i] = *plugin.NewPluginMetricType([]string{"foo", "bar"}, data)
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		enc.Encode(metrics)
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "logger: ", log.Lshortfile)
		movingAverageObj := NewMovingaverageProcessor()

		_, received_data, _ := movingAverageObj.Process("pulse.gob", buf.Bytes(), nil, logger)

		var metrics_new []plugin.PluginMetricType

		//Decodes the content into pluginMetricType
		dec := gob.NewDecoder(bytes.NewBuffer(received_data))
		dec.Decode(&metrics_new)
		So(metrics, ShouldNotResemble, metrics_new)

	})
}
