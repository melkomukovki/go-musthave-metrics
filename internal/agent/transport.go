package agent

import (
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
)

type MetricSender interface {
	SendMetrics(metrics []entities.Metric) error
}
