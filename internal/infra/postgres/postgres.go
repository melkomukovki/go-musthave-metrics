// Package postgres implements StorageService
// Using PostgreSQL to store metrics
package postgres

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
)

const (
	sqlCreateTableQuery = `
		CREATE TABLE IF NOT EXISTS metric_storage (
			name varchar(50) NOT NULL,
			type varchar(20) NOT NULL,
			value double precision NOT NULL,
			PRIMARY KEY (name, type)
		);`
	sqlAddMetricQuery     = `insert into metric_storage (name, type, value) values ($1, $2, $3) on conflict (name, type) do update set value = excluded.value;`
	sqlGetMetricQuery     = `SELECT name, type, value FROM metric_storage WHERE name=$1 AND type=$2`
	sqlGetAllMetricsQuery = `SELECT name, type, value FROM metric_storage`
)

// NewClient creates postgresql pool connection
func NewClient(connectionDSN string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), connectionDSN)
	if err != nil {
		return nil, err
	}

	err = migrate(conn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// PgRepository describes repository structure
type PgRepository struct {
	DB *pgxpool.Pool
}

// AddMetric allow to add metric to postgresql storage
func (s *PgRepository) AddMetric(ctx context.Context, metric entities.MetricInternal) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = s.retryOperation(func() error {
		_, err = s.DB.Exec(nCtx, sqlAddMetricQuery, metric.ID, metric.MType, metric.Value)
		return err
	})
	return err
}

// AddMultipleMetrics allow to add multiple metrics to postgresql storage
func (s *PgRepository) AddMultipleMetrics(ctx context.Context, metrics []entities.MetricInternal) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = s.retryOperation(func() error {
		tx, err := s.DB.Begin(nCtx)
		if err != nil {
			return err
		}
		defer func() {
			if err := tx.Rollback(nCtx); err != nil {
				log.Error().Err(err).Msg("Failed rollback transaction")
			}
		}()

		for _, metric := range metrics {
			_, err = tx.Exec(nCtx, sqlAddMetricQuery, metric.ID, metric.MType, metric.Value)
			if err != nil {
				return err
			}
		}

		err = tx.Commit(nCtx)
		return err
	})

	return err
}

// GetMetric allow to get metrics from storage
func (s *PgRepository) GetMetric(ctx context.Context, mType, mName string) (metric entities.MetricInternal, err error) {
	nCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var m entities.MetricInternal
	row := s.DB.QueryRow(nCtx, sqlGetMetricQuery, mName, mType)

	err = s.retryOperation(func() error {
		err := row.Scan(&m.ID, &m.MType, &m.Value)
		return err
	})

	if err != nil {
		return entities.MetricInternal{}, err
	}

	return m, nil
}

// GetAllMetrics allow to get all metrics from postgresql
func (s *PgRepository) GetAllMetrics(ctx context.Context) (metrics []entities.MetricInternal, err error) {
	nCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var rows pgx.Rows
	err = s.retryOperation(func() error {
		tRows, err := s.DB.Query(nCtx, sqlGetAllMetricsQuery)
		rows = tRows
		return err
	})
	if err != nil {
		return []entities.MetricInternal{}, err
	}

	for rows.Next() {
		var mSQL entities.MetricInternal

		err := rows.Scan(&mSQL.ID, &mSQL.MType, &mSQL.Value)
		if err != nil {
			return []entities.MetricInternal{}, err
		}
		metrics = append(metrics, mSQL)
	}
	return metrics, nil
}

// Ping - check connection to database
func (s *PgRepository) Ping(ctx context.Context) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return s.DB.Ping(nCtx)
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.ProtocolViolation:
			return true
		}
	}
	return false
}

func migrate(db *pgxpool.Pool) (err error) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Error().Err(err).Msg("Failed rollback transaction")
		}
	}()

	_, err = tx.Exec(ctx, sqlCreateTableQuery)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}

func (s *PgRepository) retryOperation(f func() error) error {
	const maxRetries = 3
	var retryInterval = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i := 0; i <= maxRetries; i++ {
		err := f()
		if err == nil {
			return nil
		}

		if !isRetriableError(err) {
			return err
		}

		if i < maxRetries {
			time.Sleep(retryInterval[i])
		}
	}
	return nil
}
