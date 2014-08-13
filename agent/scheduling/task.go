package scheduling

import (
	"github.com/lynxbat/pulse/agent/collection"
	"github.com/lynxbat/pulse/agent/publishing"
)

// A task definition for collecting a metric
type MetricTask struct {
	Label      string
	Metadata map[string]string
	Metrics   []collection.Metric
	Schedule  schedule
	PublisherConfig publishing.MetricPublisherConfig
	// TODO metric task stats from workers
}
