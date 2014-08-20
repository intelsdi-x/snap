package scheduling

import (
	"time"
	"github.com/lynxbat/pulse/agent/collection"
	"github.com/lynxbat/pulse/agent/publishing"
	"code.google.com/p/go-uuid/uuid"
)

// A task definition for collecting a metric
type MetricTask struct {
	Label      string
	Metadata map[string]string
	CollectorConfigs map[string]collection.CollectorConfig
	Metrics   []collection.Metric
	Schedule  schedule
	PublisherConfig publishing.MetricPublisherConfig
	Drift time.Duration

	// This is lazy loaded to avoid having to use a constructor
	// so only access via the UUID() reader method
	uuid *string

	// TODO metric task stats from workers
	// Number of times this task has fired
	tickCounter int
	lastTick time.Time
	state string
}

func (m *MetricTask) Spin(workerChan chan []collection.Metric) {
//	fmt.Printf("SPIN (%s:%s) - %d\n", m.Label, m.UUID(), m.tickCounter)

	metricMap := map[string][]collection.Metric{}

	// pivot metrics by collector type
	for _, m := range m.Metrics {
		metricMap[m.Collector] = append(metricMap[m.Collector], m)
	}

	for {
		if m.state == "stopped" {
			return // end this task go routine
		}

		// Time is referenced in this loop from this val
		t := time.Now()

		if m.Schedule.Start == nil {
			m.Trigger(workerChan, t, metricMap)
		} else if m.Schedule.Start.Before(t) || m.Schedule.Start.Equal(t) {
			m.Trigger(workerChan, t, metricMap)
		} else {
			// Sleep until start time
			time.Sleep(m.TimeTillStart())
		}
	}
}

// Returns unique ID
// This lazy loads the ID on the first load
func (m *MetricTask) UUID() string{
	if m.uuid == nil {
		u := uuid.New()
		m.uuid = &u
	}
	return *m.uuid
}

func (m *MetricTask) State() string{
	if m.state == "" {
		m.state = "waiting"
	}
	return m.state
}

func (m *MetricTask) updateDrift(t time.Time) {
	if m.lastTick.IsZero() {
		m.Drift = time.Second * 0
	} else {
		m.Drift = t.Sub(m.lastTick) - m.Schedule.Interval
	}
	m.lastTick = t
}

func (m *MetricTask) Sleep() {
	time.Sleep(m.Schedule.Interval)
}

func (m *MetricTask) Trigger(workerChan chan []collection.Metric, t time.Time, metricMap map[string][]collection.Metric) {
	m.Sleep()
	m.updateDrift(t)
	m.tickCounter++
	// Send metrics slice to worker by collector type
	for key, _ := range metricMap {
		workerChan <- metricMap[key]
	}
	m.state = "running"
	m.checkStop()
}

func (m *MetricTask) checkStop() {
	if m.Schedule.Stop != nil && (m.Schedule.Stop.Before(m.lastTick) || m.Schedule.Stop.Equal(m.lastTick)) {
		m.state = "stopped"
	}
}

func (m *MetricTask) TimeTillStart() time.Duration{
	y := m.Schedule.Start.Sub(time.Now())
	if y < 0 {
		y = 0
	}
	return y
}

func (m *MetricTask) HasStart() bool{
	return m.Schedule.Start != nil
}
