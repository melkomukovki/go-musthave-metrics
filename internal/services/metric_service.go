package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
)

// AddMetric allow to add metric
func (s *Service) AddMetric(ctx context.Context, metric entities.Metric) (err error) {
	var mName string
	var mType string
	var mValue string

	mName = metric.ID
	switch metric.MType {
	case entities.Counter:
		if metric.Delta == nil {
			return entities.ErrMissingField
		}
		mType = entities.Counter
		pMetric, err := s.GetMetric(ctx, entities.Counter, mName)
		if errors.Is(err, entities.ErrMetricNotFound) {
			mValue = strconv.Itoa(int(*metric.Delta))
		} else {
			mValue = strconv.Itoa(int(*metric.Delta + *pMetric.Delta))
		}
	case entities.Gauge:
		if metric.Value == nil {
			return entities.ErrMissingField
		}

		mType = entities.Gauge
		mValue = fmt.Sprintf("%g", *metric.Value)
	default:
		return entities.ErrMetricNotSupportedType
	}

	mSQL := entities.MetricInternal{
		ID:    mName,
		MType: mType,
		Value: mValue,
	}
	return s.ServiceRepo.AddMetric(ctx, mSQL)
}

// Ping - function to check storage availability
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.ServiceRepo.Ping(ctx)
}

// GetAllMetrics allow to get all metrics
func (s *Service) GetAllMetrics(ctx context.Context) (metrics []entities.Metric, err error) {
	mSQL, err := s.ServiceRepo.GetAllMetrics(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range mSQL {
		switch m.MType {
		case entities.Counter:
			val, err := strconv.ParseInt(m.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, entities.Metric{ID: m.ID, MType: m.MType, Delta: &val})
		case entities.Gauge:
			val, err := strconv.ParseFloat(m.Value, 64)
			if err != nil {
				return nil, err
			}
			metrics = append(metrics, entities.Metric{ID: m.ID, MType: m.MType, Value: &val})
		}
	}

	return metrics, nil
}

// GetMetric allow to get metric
func (s *Service) GetMetric(ctx context.Context, mType, mName string) (metric entities.Metric, err error) {
	m, err := s.ServiceRepo.GetMetric(ctx, mType, mName)
	if err != nil {
		return entities.Metric{}, err
	}

	switch m.MType {
	case entities.Gauge:
		val, err := strconv.ParseFloat(m.Value, 64)
		if err != nil {
			return entities.Metric{}, err
		}
		metric.ID = m.ID
		metric.MType = m.MType
		metric.Value = &val
	case entities.Counter:
		val, err := strconv.ParseInt(m.Value, 10, 64)
		if err != nil {
			return entities.Metric{}, err
		}
		metric.ID = m.ID
		metric.MType = m.MType
		metric.Delta = &val
	}
	return metric, nil
}

// AddMultipleMetrics allow to add multiple metrics
func (s *Service) AddMultipleMetrics(ctx context.Context, metrics []entities.Metric) (err error) {
	var mSQL []entities.MetricInternal
	counterMetrics := make(map[string]int64)

	for _, m := range metrics {
		switch m.MType {
		case entities.Gauge:
			if m.Value == nil {
				return entities.ErrMissingField
			}
			mSQL = append(mSQL, entities.MetricInternal{ID: m.ID, MType: m.MType, Value: fmt.Sprintf("%g", *m.Value)})
		case entities.Counter:
			if m.Delta == nil {
				return entities.ErrMissingField
			}
			counterMetrics[m.ID] += *m.Delta
		default:
			return entities.ErrMetricNotSupportedType
		}

		for metricID, aggregatedValue := range counterMetrics {
			pMetric, err := s.GetMetric(ctx, entities.Counter, metricID)
			if err != nil && !errors.Is(err, entities.ErrMetricNotFound) {
				return err
			}

			if !errors.Is(err, entities.ErrMetricNotFound) {
				aggregatedValue += *pMetric.Delta
			}

			mSQL = append(
				mSQL,
				entities.MetricInternal{ID: metricID, MType: m.MType, Value: strconv.Itoa(int(aggregatedValue))},
			)
		}
	}
	return s.ServiceRepo.AddMultipleMetrics(ctx, mSQL)
}
