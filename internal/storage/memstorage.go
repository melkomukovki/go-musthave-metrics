package storage

import (
	"errors"
	"fmt"
	"strconv"
)

// Validate implimentation
var _ Storage = MemStorage{}

type MemStorage struct {
	GaugeMetrics   map[string]float64
	CounterMetrics map[string]int64
}

func (m MemStorage) AddGaugeMetric(name, v string) error {
	vFloat, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return errors.New("can't parse value")
	}
	m.GaugeMetrics[name] = vFloat
	return nil
}

func (m MemStorage) AddCounterMetric(name, v string) error {
	vInt, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return errors.New("can't parse value")
	}
	if val, ok := m.CounterMetrics[name]; ok {
		newVal := val + vInt
		m.CounterMetrics[name] = newVal
	} else {
		m.CounterMetrics[name] = vInt
	}
	return nil
}

func (m MemStorage) GetGaugeMetric(name string) (float64, error) {
	v, ok := m.GaugeMetrics[name]
	if ok {
		return v, nil
	}
	return 0, errors.New("value not found")
}

func (m MemStorage) GetCounterMetric(name string) (int64, error) {
	v, ok := m.CounterMetrics[name]
	if ok {
		return v, nil
	}
	return 0, errors.New("value not found")
}

func (m MemStorage) GetAllMetrics() string {
	res := ""
	for k, v := range m.GaugeMetrics {
		res += fmt.Sprintf("%s:%.3f\n", k, v)
	}
	for k, v := range m.CounterMetrics {
		res += fmt.Sprintf("%s:%d\n", k, v)
	}
	return res
}
