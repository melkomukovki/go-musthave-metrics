// Package entities - entities of project
package entities

import "errors"

// Errors list
var (
	ErrMetricNotFound         = errors.New("metric not found")          // Metric not found
	ErrMetricNotSupportedType = errors.New("not supported metric type") // Unsupported metric type
	ErrMissingField           = errors.New("missing field")             // Missing required field
)
