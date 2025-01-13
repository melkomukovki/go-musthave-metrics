// Package services contains logic of application
package services

import (
	"context"

	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
)

// ServiceRepository - interface, describe storage methods
type ServiceRepository interface {
	AddMetric(ctx context.Context, metric entities.MetricInternal) (err error)
	AddMultipleMetrics(ctx context.Context, metrics []entities.MetricInternal) (err error)
	GetMetric(ctx context.Context, metricType, metricName string) (metric entities.MetricInternal, err error)
	GetAllMetrics(ctx context.Context) (metrics []entities.MetricInternal, err error)
	Ping(ctx context.Context) (err error)
}

// Service - describe service structure
type Service struct {
	ServiceRepo ServiceRepository
}
