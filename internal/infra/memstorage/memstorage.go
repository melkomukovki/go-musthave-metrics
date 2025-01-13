// Package memstorage implement ServiceStorage interface
// Use memory and filesystem to store metrics
package memstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
)

// NewClient return pointer to MemStorage structure
func NewClient(storeInterval int, storePath string, restore bool) *MemStorage {
	var syncMode = false
	if storeInterval == 0 {
		syncMode = true
	}

	newStorage := &MemStorage{
		GaugeMetrics:   sync.Map{},
		CounterMetrics: sync.Map{},
		syncStore:      syncMode,
		storeInterval:  storeInterval,
		storePath:      storePath,
	}

	if restore {
		_ = newStorage.restoreStorage()
	}

	if !newStorage.syncStore {
		go func() {
			for {
				time.Sleep(time.Duration(newStorage.storeInterval) * time.Second)
				_ = newStorage.BackupMetrics()
			}
		}()
	}

	return newStorage
}

// MemStorage define storage structure
type MemStorage struct {
	GaugeMetrics   sync.Map
	CounterMetrics sync.Map
	storeInterval  int
	syncStore      bool
	storePath      string
}

func (m *MemStorage) restoreStorage() error {
	var metrics []entities.MetricInternal
	data, err := os.ReadFile(m.storePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}
	for _, rm := range metrics {
		err = m.AddMetric(context.TODO(), rm)
		if err != nil {
			return err
		}
	}
	return nil
}

// BackupMetrics function to store metrics to filesystem
func (m *MemStorage) BackupMetrics() error {
	allMetrics, _ := m.GetAllMetrics(context.TODO())
	mJSON, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}
	return os.WriteFile(m.storePath, mJSON, 0666)
}

// AddMetric allow to add metric to storage
func (m *MemStorage) AddMetric(ctx context.Context, metric entities.MetricInternal) error {
	switch metric.MType {
	case entities.Gauge:
		if metric.Value == "" {
			return entities.ErrMissingField
		}
		tm := entities.MetricInternal{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
		m.addGaugeMetric(tm)
	case entities.Counter:
		if metric.Value == "" {
			return entities.ErrMissingField
		}
		tm := entities.MetricInternal{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
		err := m.addCounterMetric(tm)
		if err != nil {
			return err
		}
	default:
		return errors.New("not supported metric type")
	}

	if m.syncStore {
		err := m.BackupMetrics()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MemStorage) addGaugeMetric(metric entities.MetricInternal) {
	m.GaugeMetrics.Store(metric.ID, metric.Value)
}

func (m *MemStorage) addCounterMetric(metric entities.MetricInternal) (err error) {
	if val, ok := m.CounterMetrics.Load(metric.ID); ok {
		v, _ := val.(int64)
		mVal, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			return err
		}
		newVal := v + mVal
		m.CounterMetrics.Store(metric.ID, fmt.Sprintf("%d", newVal))
	} else {
		m.CounterMetrics.Store(metric.ID, metric.Value)
	}
	return nil
}

// GetMetric allow to get metric from storage
func (m *MemStorage) GetMetric(ctx context.Context, mType, mName string) (entities.MetricInternal, error) {
	switch mType {
	case entities.Gauge:
		if val, ok := m.GaugeMetrics.Load(mName); ok {
			return entities.MetricInternal{ID: mName, MType: mType, Value: val.(string)}, nil
		} else {
			return entities.MetricInternal{}, entities.ErrMetricNotFound
		}
	case entities.Counter:
		if val, ok := m.CounterMetrics.Load(mName); ok {
			return entities.MetricInternal{ID: mName, MType: mType, Value: val.(string)}, nil
		} else {
			return entities.MetricInternal{}, entities.ErrMetricNotFound
		}
	default:
		return entities.MetricInternal{}, entities.ErrMetricNotSupportedType
	}
}

// GetAllMetrics allow to get all metrics from memory storage
func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]entities.MetricInternal, error) {
	var res []entities.MetricInternal

	m.CounterMetrics.Range(func(key, value interface{}) bool {
		tm := entities.MetricInternal{
			ID:    key.(string),
			MType: entities.Counter,
			Value: value.(string),
		}
		res = append(res, tm)

		return true
	})

	m.GaugeMetrics.Range(func(key, value interface{}) bool {
		tm := entities.MetricInternal{
			ID:    key.(string),
			MType: entities.Gauge,
			Value: value.(string),
		}
		res = append(res, tm)

		return true
	})

	return res, nil
}

// Ping check accessibility. Always return nil error
func (m *MemStorage) Ping(ctx context.Context) error {
	return nil
}

// AddMultipleMetrics allow to add multiple metrics at once
func (m *MemStorage) AddMultipleMetrics(ctx context.Context, metrics []entities.MetricInternal) (err error) {
	for _, metric := range metrics {
		switch metric.MType {
		case entities.Counter:
			if metric.Value == "" {
				return entities.ErrMissingField
			}
			err := m.addCounterMetric(metric)
			if err != nil {
				return err
			}
		case entities.Gauge:
			if metric.Value == "" {
				return entities.ErrMissingField
			}
			m.addGaugeMetric(metric)
		default:
			return entities.ErrMetricNotSupportedType
		}
	}
	return nil
}
