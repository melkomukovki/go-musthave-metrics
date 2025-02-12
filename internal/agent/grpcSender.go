package agent

import (
	"context"
	"github.com/melkomukovki/go-musthave-metrics/internal/agent/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	pb "github.com/melkomukovki/go-musthave-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type GRPCMetricSender struct {
	client pb.MetricsClient
	config *config.ClientConfig
}

func NewGRPCMetricSender(cfg *config.ClientConfig) (*GRPCMetricSender, error) {
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewMetricsClient(conn)

	return &GRPCMetricSender{config: cfg, client: client}, nil
}

func (g *GRPCMetricSender) SendMetrics(metrics []entities.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	protoMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		protoMetric := &pb.Metric{
			Id:         m.ID,
			MetricType: m.MType,
		}
		switch m.MType {
		case entities.Gauge:
			protoMetric.Value = *m.Value
		case entities.Counter:
			protoMetric.Delta = *m.Delta
		}
		protoMetrics = append(protoMetrics, protoMetric)
	}

	req := &pb.AddMetricsRequest{Metrics: protoMetrics}

	_, err := g.client.AddMetrics(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
