package collection

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	collectd_cache metricCache = newMetricCache()
)

type collectDCollector struct {
	Collector
	Address string
}

type collectDConnection struct {
	connection io.ReadWriteCloser
	reader     *bufio.Reader
}

type collectDResponse struct {
	Status  int
	Message string
	Payload []string
}

type collectDConfig struct {
	Address    string
	isCaching  bool
	cachingTTL float64
}

func (c *collectDConfig) CachingEnabled() bool {
	return c.isCaching
}

func (c *collectDConfig) CacheTTL() float64 {
	return c.cachingTTL
}

// public
// addr string, caching bool, cache_ttl float64

func NewCollectDCollector(config *collectDConfig) collector {
	c := new(collectDCollector)
	c.Address = config.Address
	c.Caching = config.CachingEnabled()
	c.CachingTTL = config.CacheTTL()
	return c
}

func (c *collectDCollector) GetMetricList() []Metric {
	conn := c.newConnection()
	conn.sendCommand("LISTVAL")
	resp := conn.readResponse()

	metrics := make([]Metric, len(resp.Payload))
	for x := 0; x < len(resp.Payload); x++ {
		a := strings.Split(resp.Payload[x], " ")
		c := strings.Split(a[1], "/")
		host := c[0]
		namespace := c[1:]

		var ms_str string
		ms_a := strings.Split(a[0], ".")

		if len(ms_a) > 1 {
			ms_str = ms_a[0] + ms_a[1]
		} else {
			ms_str = ms_a[0]
		}
		update_time, _ := msToTime(ms_str)

		metrics[x] = Metric{host, namespace, update_time, map[string]metricType{}, "collectd", Polling}
	}

	defer conn.Close()
	return metrics
}

func (c *collectDCollector) pullMetrics(conn *collectDConnection, in chan Metric, out chan Metric, doquit chan bool) {
	for {
		select {
		case metric := <-in:
			// We received a metric
			conn.sendCommand("GETVAL " + metric.GetFullNamespace())
			resp := conn.readResponse()
			for x := 0; x < len(resp.Payload); x++ {
				a := strings.Split(resp.Payload[x], "=")
				name := a[0]
				value, _ := strconv.ParseFloat(a[1], 64)
				metric.Values[name] = value
			}
			out <- metric
		case <-doquit:
			return
		}
	}
}

func (c *collectDCollector) GetMetricValues(metrics []Metric, things ...interface{}) []Metric {
	if !c.Caching || (c.Caching && collectd_cache.IsExpired(c.CachingTTL)) {
		var cores int

		if len(things) > 0 {
			cores = things[0].(int)
		} else {
			cores = 1
		}

		connections := make([]*collectDConnection, cores)

		// Our metric passing channels
		complete_chan := make(chan Metric)
		todo_chan := make(chan Metric, len(metrics))
		// Our quit channel
		quit_chan := make(chan bool)

		for x := 0; x < cores; x++ {
			connections[x] = c.newConnection()
			go c.pullMetrics(connections[x], todo_chan, complete_chan, quit_chan)
		}

		for _, m := range metrics {
			todo_chan <- m
		}

		out_metrics := []Metric{}
		for len(out_metrics) < len(metrics) {
			select {
			case m := <-complete_chan:
				out_metrics = append(out_metrics, m)
			}
		}

		// Signal all channels to close
		for x := 0; x < cores; x++ {
			quit_chan <- true
		}
		// Close all connections
		for _, conn := range connections {
			conn.Close()
		}
		collectd_cache.Metrics = out_metrics
		collectd_cache.LastPull = time.Now()
		collectd_cache.New = false
	}
	return collectd_cache.Metrics
}

func (c *collectDCollector) newConnection() *collectDConnection {
	conn, err := net.Dial("unix", c.Address)
	if err != nil {
		panic(err.Error())
	}
	r := bufio.NewReader(conn)
	return &collectDConnection{conn, r}
}

func (c *collectDConnection) Close() {
	c.connection.Close()
}

func (c *collectDConnection) sendCommand(cmd string) {
	_, err := c.connection.Write([]byte(cmd + "\n"))
	if err != nil {
		panic(err.Error())
	}
}

func (c *collectDConnection) readResponse() *collectDResponse {
	resp := new(collectDResponse)
	// Read out status
	_, err := fmt.Fscanf(c.reader, "%d ", &resp.Status)
	if err != nil {
		panic(err.Error())
	}
	// Read response message
	resp.Message, err = c.reader.ReadString('\n')
	if err != nil {
		panic(err.Error())
	}
	// Read payload
	return resp.parsePayload(c.reader)
}

func (resp *collectDResponse) parsePayload(r *bufio.Reader) *collectDResponse {
	resp.Payload = make([]string, resp.Status)
	for x := 0; x < resp.Status; x++ {
		s, _ := r.ReadString('\n')
		resp.Payload[x] = s[:len(s)-1]
	}
	return resp
}
