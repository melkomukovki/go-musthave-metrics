package storage

import (
	"encoding/json"
	"errors"
	"os"
)

// Validate implimentation
var _ Storage = MemStorage{}

func NewMemStorage(storeInterval int, storePath string) *MemStorage {
	var sStore = false
	if storeInterval == 0 {
		sStore = true
	}
	return &MemStorage{
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
		SyncStore:      sStore,
		storeInterval:  storeInterval,
		StorePath:      storePath,
	}
}

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
	storeInterval  int
	SyncStore      bool
	StorePath      string
}

func (m MemStorage) RestoreStorage() error {
	metrics := []Metrics{}
	data, err := os.ReadFile(m.StorePath)
	if err != nil {
		return err
	}
	json.Unmarshal(data, &metrics)
	for _, rm := range metrics {
		m.AddMetric(rm)
	}
	return nil
}

func (m MemStorage) BackupMetrics() error {
	allMetrics := m.GetAllMetrics()
	mJSON, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}
	os.WriteFile(m.StorePath, mJSON, 0666)
	return nil
}

func (m MemStorage) AddMetric(metric Metrics) error {
	switch metric.MType {
	case Gauge:
		if metric.Value == nil {
			return errors.New("field 'value' can't be empty for metric with type gauge")
		}
		tm := GaugeMetrics{
			ID:    metric.ID,
			MType: metric.MType,
			Value: metric.Value,
		}
		m.addGaugeMetric(&tm)
	case Counter:
		if metric.Delta == nil {
			return errors.New("field 'delta' can't be empty for metric with type counter ")
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

	if m.SyncStore {
		m.BackupMetrics()
	}
	return nil
}

func (m MemStorage) addGaugeMetric(metric *GaugeMetrics) {
	m.GaugeMetrics[metric.ID] = *metric.Value
}

func (m MemStorage) addCounterMetric(metric *CounterMetrics) {
	if val, ok := m.CounterMetrics[metric.ID]; ok {
		newVal := val + *metric.Delta
		m.CounterMetrics[metric.ID] = newVal
	} else {
		m.CounterMetrics[metric.ID] = *metric.Delta
	}
}

func (m MemStorage) GetMetric(mType, mName string) (Metrics, error) {
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

func (m MemStorage) GetAllMetrics() []Metrics {
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
	return res
}
