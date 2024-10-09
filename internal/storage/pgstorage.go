package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	_ Storage = &PgStorage{}
)

type PgStorage struct {
	SyncStore bool
	dbConn    *pgx.Conn
}

func NewPgStorage(connectionDSN string) (*PgStorage, error) {
	conn, err := pgx.Connect(context.Background(), connectionDSN)
	if err != nil {
		return nil, err
	}

	return &PgStorage{dbConn: conn}, nil
}

func (p *PgStorage) SyncStorage() (flag bool) {
	return false
}

func (p *PgStorage) AddMetric(metric Metrics) (err error) {
	return nil
}

func (p *PgStorage) GetMetric(metricType, metricName string) (metric Metrics, err error) {
	return Metrics{}, nil
}

func (p *PgStorage) GetAllMetrics() (metrics []Metrics) {
	return []Metrics{}
}

func (p *PgStorage) RestoreStorage() (err error) {
	return nil
}

func (p *PgStorage) BackupMetrics() (err error) {
	return nil
}

func (p *PgStorage) Ping() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
	defer cancel()
	return p.dbConn.Ping(ctx)
}
