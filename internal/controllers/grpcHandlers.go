package controllers

import (
	"context"
	"fmt"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	pb "github.com/melkomukovki/go-musthave-metrics/internal/proto"
	"github.com/melkomukovki/go-musthave-metrics/internal/services"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	service *services.Service
}

func NewMetricsServer(service *services.Service) *MetricsServer {
	return &MetricsServer{service: service}
}

func (s *MetricsServer) AddMetric(ctx context.Context, req *pb.AddMetricRequest) (*pb.AddMetricResponse, error) {
	metric := entities.Metric{
		ID:    req.Metric.Id,
		MType: req.Metric.MetricType,
	}
	switch req.Metric.MetricType {
	case entities.Gauge:
		metric.Value = &req.Metric.Value
	case entities.Counter:
		metric.Delta = &req.Metric.Delta
	default:
		return nil, fmt.Errorf("unknown metric type: %s", req.Metric.MetricType)
	}

	err := s.service.AddMetric(ctx, metric)
	if err != nil {
		return nil, err
	}

	return &pb.AddMetricResponse{Message: "Success"}, nil
}

func (s *MetricsServer) AddMetrics(ctx context.Context, req *pb.AddMetricsRequest) (*pb.AddMetricsResponse, error) {
	var metrics []entities.Metric
	for _, m := range req.Metrics {
		metric := entities.Metric{
			ID:    m.Id,
			MType: m.MetricType,
		}
		switch m.MetricType {
		case entities.Gauge:
			metric.Value = &m.Value
		case entities.Counter:
			metric.Delta = &m.Delta
		default:
			return nil, fmt.Errorf("unknown metric type: %s", m.MetricType)
		}
		metrics = append(metrics, metric)
	}

	err := s.service.AddMultipleMetrics(ctx, metrics)
	if err != nil {
		return nil, err
	}

	return &pb.AddMetricsResponse{Message: "Success"}, nil
}

func (s *MetricsServer) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	metric, err := s.service.GetMetric(ctx, req.MetricType, req.Id)
	if err != nil {
		return nil, err
	}

	respMetric := pb.Metric{
		Id:         metric.ID,
		MetricType: metric.MType,
	}
	if metric.MType == entities.Gauge {
		respMetric.Value = *metric.Value
	} else if metric.MType == entities.Counter {
		respMetric.Delta = *metric.Delta
	}

	return &pb.GetMetricResponse{Metric: &respMetric}, nil
}

func (s *MetricsServer) ListMetrics(ctx context.Context, req *pb.ListMetricsRequest) (*pb.ListMetricsResponse, error) {
	metrics, err := s.service.GetAllMetrics(ctx)
	if err != nil {
		return nil, err
	}

	var pbMetrics []*pb.Metric
	for _, m := range metrics {
		pbMetric := &pb.Metric{
			Id:         m.ID,
			MetricType: m.MType,
		}
		if m.MType == entities.Gauge {
			pbMetric.Value = *m.Value
		} else if m.MType == entities.Counter {
			pbMetric.Delta = *m.Delta
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}

	return &pb.ListMetricsResponse{Metrics: pbMetrics}, nil
}

func (s *MetricsServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.service.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}
	return &pb.PingResponse{Message: "Success"}, nil
}
