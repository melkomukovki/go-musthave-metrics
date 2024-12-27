package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestService_AddMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockServiceRepository(ctrl)
	s := &Service{ServiceRepo: mockRepo}

	tests := []struct {
		name          string
		input         entities.Metric
		setupMock     func()
		expectedError error
	}{
		{
			name: "Add Gauge metric",
			input: entities.Metric{
				ID:    "gaugeMetric",
				MType: entities.Gauge,
				Value: func(v float64) *float64 { return &v }(123.456),
			},
			setupMock: func() {
				mockRepo.EXPECT().AddMetric(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Add Counter metric",
			input: entities.Metric{
				ID:    "counterMetric",
				MType: entities.Counter,
				Delta: func(v int64) *int64 { return &v }(10),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetMetric(gomock.Any(), entities.Counter, "counterMetric").
					Return(entities.MetricInternal{}, entities.ErrMetricNotFound)
				mockRepo.EXPECT().
					AddMetric(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Add Counter metric with aggregation",
			input: entities.Metric{
				ID:    "counterMetric",
				MType: entities.Counter,
				Delta: func(v int64) *int64 { return &v }(10),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetMetric(gomock.Any(), entities.Counter, "counterMetric").
					Return(entities.MetricInternal{
						ID:    "counterMetric",
						MType: entities.Counter,
						Value: "20",
					}, nil)
				mockRepo.EXPECT().
					AddMetric(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Missing field for Gauge",
			input: entities.Metric{
				ID:    "gauge_metric",
				MType: entities.Gauge,
			},
			setupMock:     func() {},
			expectedError: entities.ErrMissingField,
		},
		{
			name: "Unsupported metric type",
			input: entities.Metric{
				ID:    "unknownMetric",
				MType: "nonType",
			},
			setupMock:     func() {},
			expectedError: entities.ErrMetricNotSupportedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := s.AddMetric(context.Background(), tt.input)

			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestService_GetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockServiceRepository(ctrl)
	s := &Service{ServiceRepo: mockRepo}

	tests := []struct {
		name           string
		mType          string
		mName          string
		setupMock      func()
		expectedMetric entities.Metric
		expectedError  error
	}{
		{
			name:  "Get Gauge metric",
			mType: entities.Gauge,
			mName: "gaugeMetric",
			setupMock: func() {
				mockRepo.EXPECT().
					GetMetric(gomock.Any(), entities.Gauge, "gaugeMetric").
					Return(entities.MetricInternal{
						ID:    "gaugeMetric",
						MType: entities.Gauge,
						Value: "123.456",
					}, nil)
			},
			expectedMetric: entities.Metric{
				ID:    "gaugeMetric",
				MType: entities.Gauge,
				Value: func(v float64) *float64 { return &v }(123.456),
			},
			expectedError: nil,
		},
		{
			name:  "Get Counter metric",
			mType: entities.Counter,
			mName: "counterMetric",
			setupMock: func() {
				mockRepo.EXPECT().
					GetMetric(gomock.Any(), entities.Counter, "counterMetric").
					Return(entities.MetricInternal{
						ID:    "counterMetric",
						MType: entities.Counter,
						Value: "10",
					}, nil)
			},
			expectedMetric: entities.Metric{
				ID:    "counterMetric",
				MType: entities.Counter,
				Delta: func(v int64) *int64 { return &v }(10),
			},
			expectedError: nil,
		},
		{
			name:  "Metric not found",
			mType: entities.Counter,
			mName: "unknownMetric",
			setupMock: func() {
				mockRepo.EXPECT().
					GetMetric(gomock.Any(), entities.Counter, "unknownMetric").
					Return(entities.MetricInternal{}, entities.ErrMetricNotFound)
			},
			expectedMetric: entities.Metric{},
			expectedError:  entities.ErrMetricNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			metric, err := s.GetMetric(context.Background(), tt.mType, tt.mName)

			assert.Equal(t, tt.expectedMetric, metric)
			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestService_GetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockServiceRepository(ctrl)
	s := &Service{ServiceRepo: mockRepo}

	mockMetrics := []entities.MetricInternal{
		{ID: "gaugeMetric", MType: entities.Gauge, Value: "123.456"},
		{ID: "counterMetric", MType: entities.Counter, Value: "10"},
	}

	expectedMetrics := []entities.Metric{
		{ID: "gaugeMetric", MType: entities.Gauge, Value: func(v float64) *float64 { return &v }(123.456)},
		{ID: "counterMetric", MType: entities.Counter, Delta: func(v int64) *int64 { return &v }(10)},
	}

	mockRepo.EXPECT().
		GetAllMetrics(gomock.Any()).
		Return(mockMetrics, nil)

	metrics, err := s.GetAllMetrics(context.Background())

	assert.Equal(t, expectedMetrics, metrics)
	assert.NoError(t, err)
}

func TestService_AddMultipleMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockServiceRepository(ctrl)
	s := &Service{ServiceRepo: mockRepo}

	tests := []struct {
		name          string
		input         []entities.Metric
		setupMock     func()
		expectedError error
	}{
		{
			name: "Add multiple metrics",
			input: []entities.Metric{
				{ID: "gaugeMetric", MType: entities.Gauge, Value: func(v float64) *float64 { return &v }(123.456)},
				{ID: "counterMetric", MType: entities.Counter, Delta: func(v int64) *int64 { return &v }(10)},
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetMetric(gomock.Any(), entities.Counter, "counterMetric").
					Return(entities.MetricInternal{}, entities.ErrMetricNotFound)
				mockRepo.EXPECT().
					AddMultipleMetrics(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Missing field in Gauge metric",
			input: []entities.Metric{
				{ID: "gaugeMetric", MType: entities.Gauge},
			},
			setupMock:     func() {},
			expectedError: entities.ErrMissingField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := s.AddMultipleMetrics(context.Background(), tt.input)

			assert.ErrorIs(t, err, tt.expectedError)
		})
	}
}

func TestService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockServiceRepository(ctrl)
	s := &Service{ServiceRepo: mockRepo}

	mockRepo.EXPECT().
		Ping(gomock.Any()).
		Return(nil)

	err := s.Ping(context.Background())

	assert.NoError(t, err)
}
