package model

type MetricType int

const (
	Counter MetricType = iota
	Gauge
)

type Metric struct {
	ID    string      `json:"id,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Type  MetricType  `json:"-"`
}
