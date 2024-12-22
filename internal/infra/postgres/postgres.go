package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/melkomukovki/go-musthave-metrics/internal/entities"
	"time"
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

type PgRepository struct {
	DB *pgxpool.Pool
}

func (s *PgRepository) AddMetric(ctx context.Context, metric entities.MetricsSQL) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(2*time.Second))
	defer cancel()

	err = s.retryOperation(func() error {
		_, err = s.DB.Exec(nCtx, sqlAddMetricQuery, metric.ID, metric.MType, metric.Value)
		return err
	})
	return err
}

func (s *PgRepository) AddMultipleMetrics(ctx context.Context, metrics []entities.MetricsSQL) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(3*time.Second))
	defer cancel()

	err = s.retryOperation(func() error {
		tx, err := s.DB.Begin(nCtx)
		if err != nil {
			return err
		}
		defer tx.Rollback(nCtx)

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

func (s *PgRepository) GetMetric(ctx context.Context, mType, mName string) (metric entities.MetricsSQL, err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(2*time.Second))
	defer cancel()

	var m entities.MetricsSQL
	row := s.DB.QueryRow(nCtx, sqlGetMetricQuery, mName, mType)

	err = s.retryOperation(func() error {
		err := row.Scan(&m.ID, &m.MType, &m.Value)
		return err
	})

	if err != nil {
		return entities.MetricsSQL{}, err
	}

	return m, nil
}

func (s *PgRepository) GetAllMetrics(ctx context.Context) (metrics []entities.MetricsSQL, err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(3*time.Second))
	defer cancel()

	var rows pgx.Rows
	err = s.retryOperation(func() error {
		tRows, err := s.DB.Query(nCtx, sqlGetAllMetricsQuery)
		rows = tRows
		return err
	})
	if err != nil {
		return []entities.MetricsSQL{}, err
	}

	for rows.Next() {
		var mSQL entities.MetricsSQL

		err := rows.Scan(&mSQL.ID, &mSQL.MType, &mSQL.Value)
		if err != nil {
			return []entities.MetricsSQL{}, err
		}
		metrics = append(metrics, mSQL)
	}
	return metrics, nil
}

func (s *PgRepository) Ping(ctx context.Context) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(time.Second))
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
	defer tx.Rollback(ctx)

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
