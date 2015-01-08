package availability

import (
	"sort"

	"github.com/intelsdilabs/pulse/metric"
)

type AvailablePlugin struct {
	Metrics *metrics
}

func (ap *AvailablePlugin) Subscriptions() int {
	return ap.Metrics.Max()
}

type metrics map[string]*metric.Metric

func (m *metrics) Max() int {
	if m.len() > 0 {
		var subs []int
		for _, s := range *m {
			subs = append(subs, s.Subscriptions())
		}
		sort.Ints(subs)
		return subs[len(subs)-1]
	}
	return 0
}

func (m *metrics) len() int {
	return len(*m)
}
