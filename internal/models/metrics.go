package models

import "gmetrics/internal/metrics"

type GaugeMetric struct {
	id    string
	value metrics.Gauge
}

type CounterMetric struct {
	id    string
	value metrics.Gauge
}
