package entities

// Metric types
const (
	Gauge   = "gauge"
	Counter = "counter"
)

// Metric define model for external usage
type Metric struct {
	ID    string   `json:"id" binding:"required"`   // Metric name
	MType string   `json:"type" binding:"required"` // Metric type
	Delta *int64   `json:"delta,omitempty"`         // Value for counter metric
	Value *float64 `json:"value,omitempty"`         // Value for gauge metric
}

// MetricInternal define model for internal usage
type MetricInternal struct {
	ID    string
	MType string
	Value string
}
