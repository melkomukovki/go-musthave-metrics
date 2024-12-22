package services

import (
	"context"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
)

type ServiceRepository interface {
	AddMetric(ctx context.Context, metric entities.MetricsSQL) (err error)
	AddMultipleMetrics(ctx context.Context, metrics []entities.MetricsSQL) (err error)
	GetMetric(ctx context.Context, metricType, metricName string) (metric entities.MetricsSQL, err error)
	GetAllMetrics(ctx context.Context) (metrics []entities.MetricsSQL, err error)
	Ping(ctx context.Context) (err error)
}

type Service struct {
	ServiceRepo ServiceRepository
}
