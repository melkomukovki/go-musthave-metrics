package storage

import (
	"context"
	"errors"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

var (
	ErrMetricNotFound         = errors.New("metric not found")
	ErrMetricNotSupportedType = errors.New("not supported metric type")
	ErrMissingField           = errors.New("missing field")
)

type Storage interface {
	AddMetric(ctx context.Context, metric Metrics) (err error)
	AddMultipleMetrics(ctx context.Context, metrics []Metrics) (err error)
	GetMetric(ctx context.Context, metricType, metricName string) (metric Metrics, err error)
	GetAllMetrics(ctx context.Context) (metrics []Metrics, err error)
	Ping(ctx context.Context) (err error)
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type GaugeMetrics struct {
	ID    string
	MType string
	Value *float64
}

type CounterMetrics struct {
	ID    string
	MType string
	Delta *int64
}
