package main

import (
	"errors"
	"net/http"
	"strconv"
)

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

type Storage interface {
	AddGaugeMetric(string, string) error
	AddCounterMetric(string, string) error
}

var storage Storage = MemStorage{
	GaugeMetrics:   make(map[string]float64),
	CounterMetrics: make(map[string]int64),
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST requests", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-type must be 'text/plain'", http.StatusBadRequest)
		return
	}

	mType := r.PathValue("metricType")
	mName := r.PathValue("metricName")
	mValue := r.PathValue("metricValue")

	if mType == "gauge" {
		err := storage.AddGaugeMetric(mName, mValue)
		if err != nil {
			http.Error(w, "Cant add gauge metric", http.StatusBadRequest)
			return
		}
	} else if mType == "counter" {
		err := storage.AddCounterMetric(mName, mValue)
		if err != nil {
			http.Error(w, "Cant add counter metric", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Unknown metric type. Use gauge or counter.", http.StatusBadRequest)
		return
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", metricsHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad request", http.StatusBadRequest)
	})

	err := http.ListenAndServe("0.0.0.0:8080", mux)
	if err != nil {
		panic(err)
	}
}
