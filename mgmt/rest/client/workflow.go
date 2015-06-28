package client

type Workflow struct {
	MTs        []*MetricType `json:"metric_types"`
	Publishers []string      `json:"publishers"`
}
