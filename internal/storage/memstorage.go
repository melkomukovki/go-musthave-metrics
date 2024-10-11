package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
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
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
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
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
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
	m.GaugeMetrics[metric.ID] = *metric.Value
}

func (m *MemStorage) addCounterMetric(metric *CounterMetrics) {
	if val, ok := m.CounterMetrics[metric.ID]; ok {
		newVal := val + *metric.Delta
		m.CounterMetrics[metric.ID] = newVal
	} else {
		m.CounterMetrics[metric.ID] = *metric.Delta
	}
}

func (m *MemStorage) GetMetric(ctx context.Context, mType, mName string) (Metrics, error) {
	switch mType {
	case Gauge:
		if val, ok := m.GaugeMetrics[mName]; ok {
			tm := Metrics{
				ID:    mName,
				MType: Gauge,
				Value: &val,
			}
			return tm, nil
		} else {
			return Metrics{}, errors.New("metric not found")
		}
	case Counter:
		if val, ok := m.CounterMetrics[mName]; ok {
			tm := Metrics{
				ID:    mName,
				MType: Counter,
				Delta: &val,
			}
			return tm, nil
		} else {
			return Metrics{}, errors.New("metric not found")
		}
	default:
		return Metrics{}, errors.New("not supported metric type")
	}
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]Metrics, error) {
	var res []Metrics
	for k, v := range m.CounterMetrics {
		tm := Metrics{
			ID:    k,
			MType: Counter,
			Delta: &v,
		}
		res = append(res, tm)
	}
	for k, v := range m.GaugeMetrics {
		tm := Metrics{
			ID:    k,
			MType: Gauge,
			Value: &v,
		}
		res = append(res, tm)
	}
	return res, nil
}

func (m *MemStorage) Ping(ctx context.Context) error {
	return nil
}
