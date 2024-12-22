package memstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	"os"
	"strconv"
	"sync"
	"time"
)

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

type MemStorage struct {
	GaugeMetrics   sync.Map
	CounterMetrics sync.Map
	storeInterval  int
	syncStore      bool
	storePath      string
}

func (m *MemStorage) restoreStorage() error {
	metrics := []entities.MetricsSQL{}
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

func (m *MemStorage) BackupMetrics() error {
	allMetrics, _ := m.GetAllMetrics(context.TODO())
	mJSON, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}
	return os.WriteFile(m.storePath, mJSON, 0666)
}

func (m *MemStorage) AddMetric(ctx context.Context, metric entities.MetricsSQL) error {
	switch metric.MType {
	case entities.Gauge:
		if metric.Value == "" {
			return entities.ErrMissingField
		}
		tm := entities.MetricsSQL{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
		m.addGaugeMetric(tm)
	case entities.Counter:
		if metric.Value == "" {
			return entities.ErrMissingField
		}
		tm := entities.MetricsSQL{
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

func (m *MemStorage) addGaugeMetric(metric entities.MetricsSQL) {
	m.GaugeMetrics.Store(metric.ID, metric.Value)
}

func (m *MemStorage) addCounterMetric(metric entities.MetricsSQL) (err error) {
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

func (m *MemStorage) GetMetric(ctx context.Context, mType, mName string) (entities.MetricsSQL, error) {
	switch mType {
	case entities.Gauge:
		if val, ok := m.GaugeMetrics.Load(mName); ok {
			return entities.MetricsSQL{ID: mName, MType: mType, Value: val.(string)}, nil
		} else {
			return entities.MetricsSQL{}, entities.ErrMetricNotFound
		}
	case entities.Counter:
		if val, ok := m.CounterMetrics.Load(mName); ok {
			return entities.MetricsSQL{ID: mName, MType: mType, Value: val.(string)}, nil
		} else {
			return entities.MetricsSQL{}, entities.ErrMetricNotFound
		}
	default:
		return entities.MetricsSQL{}, entities.ErrMetricNotSupportedType
	}
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]entities.MetricsSQL, error) {
	var res []entities.MetricsSQL

	m.CounterMetrics.Range(func(key, value interface{}) bool {
		tm := entities.MetricsSQL{
			ID:    key.(string),
			MType: entities.Counter,
			Value: value.(string),
		}
		res = append(res, tm)

		return true
	})

	m.GaugeMetrics.Range(func(key, value interface{}) bool {
		tm := entities.MetricsSQL{
			ID:    key.(string),
			MType: entities.Gauge,
			Value: value.(string),
		}
		res = append(res, tm)

		return true
	})

	return res, nil
}

func (m *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MemStorage) AddMultipleMetrics(ctx context.Context, metrics []entities.MetricsSQL) (err error) {
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
