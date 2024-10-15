package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

var (
	_ Storage = &MemStorage{}
)

func NewMemStorage(storeInterval int, storePath string, restore bool) *MemStorage {
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
		newStorage.restoreStorage()
	}

	if !newStorage.syncStore {
		go func() {
			for {
				time.Sleep(time.Duration(newStorage.storeInterval) * time.Second)
				newStorage.BackupMetrics()
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
	metrics := []Metrics{}
	data, err := os.ReadFile(m.storePath)
	if err != nil {
		return err
	}
	json.Unmarshal(data, &metrics)
	for _, rm := range metrics {
		m.AddMetric(context.TODO(), rm)
	}
	return nil
}

func (m *MemStorage) BackupMetrics() error {
	allMetrics, _ := m.GetAllMetrics(context.TODO())
	mJSON, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}
	os.WriteFile(m.storePath, mJSON, 0666)
	return nil
}

func (m *MemStorage) AddMetric(ctx context.Context, metric Metrics) error {
	switch metric.MType {
	case Gauge:
		if metric.Value == nil {
			return ErrMissingField
		}
		tm := GaugeMetrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
		m.addGaugeMetric(&tm)
	case Counter:
		if metric.Delta == nil {
			return ErrMissingField
		}
		tm := CounterMetrics{
			ID:    metric.ID,
			MType: metric.MType,
			Delta: metric.Delta,
		}
		m.addCounterMetric(&tm)
	default:
		return errors.New("not supported metric type")
	}

	if m.syncStore {
		m.BackupMetrics()
	}
	return nil
}

func (m *MemStorage) addGaugeMetric(metric *GaugeMetrics) {
	m.GaugeMetrics.Store(metric.ID, *metric.Value)
}

func (m *MemStorage) addCounterMetric(metric *CounterMetrics) {
	if val, ok := m.CounterMetrics.Load(metric.ID); ok {
		v, _ := val.(int64)
		newVal := v + *metric.Delta
		m.CounterMetrics.Store(metric.ID, newVal)
	} else {
		m.CounterMetrics.Store(metric.ID, *metric.Delta)
	}
}

func (m *MemStorage) GetMetric(ctx context.Context, mType, mName string) (Metrics, error) {
	switch mType {
	case Gauge:
		if val, ok := m.GaugeMetrics.Load(mName); ok {
			v, ok := val.(float64)
			if !ok {
				return Metrics{}, ErrWrongValue
			}
			tm := Metrics{
				ID:    mName,
				MType: Gauge,
				Value: &v,
			}
			return tm, nil
		} else {
			return Metrics{}, ErrMetricNotFound
		}
	case Counter:
		if val, ok := m.CounterMetrics.Load(mName); ok {
			v, ok := val.(int64)
			if !ok {
				return Metrics{}, ErrWrongValue
			}
			tm := Metrics{
				ID:    mName,
				MType: Counter,
				Delta: &v,
			}
			return tm, nil
		} else {
			return Metrics{}, ErrMetricNotFound
		}
	default:
		return Metrics{}, ErrMetricNotSupportedType
	}
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]Metrics, error) {
	var res []Metrics

	m.CounterMetrics.Range(func(key, value interface{}) bool {
		val, ok := value.(int64)
		if !ok {
			return false
		}
		tm := Metrics{
			ID:    key.(string),
			MType: Counter,
			Delta: &val,
		}
		res = append(res, tm)

		return true
	})

	m.GaugeMetrics.Range(func(key, value interface{}) bool {
		val, ok := value.(float64)
		if !ok {
			return false
		}
		tm := Metrics{
			ID:    key.(string),
			MType: Gauge,
			Value: &val,
		}
		res = append(res, tm)

		return true
	})

	return res, nil
}

func (m *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (m *MemStorage) AddMultipleMetrics(ctx context.Context, metrics []Metrics) (err error) {
	for _, metric := range metrics {
		switch metric.MType {
		case Counter:
			if metric.Delta == nil {
				return ErrMissingField
			}
			tm := CounterMetrics{
				ID:    metric.ID,
				MType: Counter,
				Delta: metric.Delta,
			}
			m.addCounterMetric(&tm)
		case Gauge:
			if metric.Value == nil {
				return ErrMissingField
			}
			tm := GaugeMetrics{
				ID:    metric.ID,
				MType: Gauge,
				Value: metric.Value,
			}
			m.addGaugeMetric(&tm)
		default:
			return ErrMetricNotSupportedType
		}
	}
	return nil
}
