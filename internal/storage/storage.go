package storage

type Storage interface {
	AddGaugeMetric(string, string) error
	AddCounterMetric(string, string) error
	GetGaugeMetric(string) (float64, error)
	GetCounterMetric(string) (int64, error)
	GetAllMetrics() string
}
