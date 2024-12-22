package entities

import "errors"

var (
	ErrMetricNotFound         = errors.New("metric not found")
	ErrMetricNotSupportedType = errors.New("not supported metric type")
	ErrMissingField           = errors.New("missing field")
	ErrWrongValue             = errors.New("wrong metric value")
)
