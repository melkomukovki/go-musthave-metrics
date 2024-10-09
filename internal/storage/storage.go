package storage

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Storage interface {
	AddMetric(metric Metrics) (err error)
	GetMetric(metricType, metricName string) (metric Metrics, err error)
	GetAllMetrics() (metrics []Metrics)
	RestoreStorage() (err error)
	BackupMetrics() (err error)
	SyncStorage() (flag bool)
	Ping() (err error)
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type GaugeMetrics struct {
	ID    string
	MType string
	Value *float64
}

type CounterMetrics struct {
	ID    string
	MType string
	Delta *int64
}
