package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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

var (
	_ Storage = &PgStorage{}
)

type PgStorage struct {
	SyncStore bool
	dbPool    *pgxpool.Pool
}

func NewPgStorage(connectionDSN string) (*PgStorage, error) {
	conn, err := pgxpool.New(context.Background(), connectionDSN)
	if err != nil {
		return nil, err
	}

	pgStorage := &PgStorage{dbPool: conn}
	err = pgStorage.migrate()
	if err != nil {
		return nil, fmt.Errorf("can't make migrations. Error: %s", err.Error())
	}

	return pgStorage, nil
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

func (p *PgStorage) migrate() (err error) {
	ctx := context.Background()
	tx, err := p.dbPool.Begin(ctx)
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

func (p *PgStorage) retryOperation(f func() error) error {
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

func (p *PgStorage) AddMetric(ctx context.Context, metric Metrics) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(2*time.Second))
	defer cancel()

	var mName string
	var mType string
	var mValue string

	mName = metric.ID
	if metric.MType == Counter {
		if metric.Delta == nil {
			return ErrMissingField
		}

		mType = Counter
		pMetric, err := p.GetMetric(nCtx, Counter, mName)
		if errors.Is(err, ErrMetricNotFound) {
			mValue = strconv.Itoa(int(*metric.Delta))
		} else {
			mValue = strconv.Itoa(int(*metric.Delta + *pMetric.Delta))
		}
	} else if metric.MType == Gauge {
		if metric.Value == nil {
			return ErrMissingField
		}

		mType = Gauge
		mValue = fmt.Sprintf("%g", *metric.Value)
	} else {
		return ErrMetricNotSupportedType
	}

	err = p.retryOperation(func() error {
		_, err = p.dbPool.Exec(nCtx, sqlAddMetricQuery, mName, mType, mValue)
		return err
	})
	return err
}

func (p *PgStorage) GetMetric(ctx context.Context, metricType, metricName string) (metric Metrics, err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(2*time.Second))
	defer cancel()

	var m Metrics
	row := p.dbPool.QueryRow(nCtx, sqlGetMetricQuery, metricName, metricType)
	switch metricType {
	case Gauge:
		err := p.retryOperation(func() error {
			err := row.Scan(&m.ID, &m.MType, &m.Value)
			return err
		})
		if err != nil {
			return Metrics{}, ErrMetricNotFound
		}
	case Counter:
		err := p.retryOperation(func() error {
			err := row.Scan(&m.ID, &m.MType, &m.Delta)
			return err
		})
		if err != nil {
			return Metrics{}, ErrMetricNotFound
		}
	}
	return m, nil
}

func (p *PgStorage) GetAllMetrics(ctx context.Context) (metrics []Metrics, err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(3*time.Second))
	defer cancel()

	var rows pgx.Rows
	err = p.retryOperation(func() error {
		tRows, err := p.dbPool.Query(nCtx, sqlGetAllMetricsQuery)
		rows = tRows
		return err
	})
	if err != nil {
		return []Metrics{}, err
	}

	for rows.Next() {
		var mName string
		var mType string
		var mValue string

		err := rows.Scan(&mName, &mType, &mValue)
		if err != nil {
			return []Metrics{}, err
		}
		if mType == Counter {
			val, err := strconv.ParseInt(mValue, 10, 64)
			if err != nil {
				return []Metrics{}, err
			}
			metrics = append(metrics, Metrics{ID: mName, MType: Counter, Delta: &val})
		} else if mType == Gauge {
			val, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				return []Metrics{}, err
			}
			metrics = append(metrics, Metrics{ID: mName, MType: Gauge, Value: &val})
		}
	}
	return metrics, nil
}

func (p *PgStorage) Ping(ctx context.Context) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(time.Second))
	defer cancel()
	return p.dbPool.Ping(nCtx)
}

func (p *PgStorage) AddMultipleMetrics(ctx context.Context, metrics []Metrics) (err error) {
	nCtx, cancel := context.WithTimeout(ctx, time.Duration(3*time.Second))
	defer cancel()

	counterMetrics := make(map[string]int64)

	err = p.retryOperation(func() error {
		tx, err := p.dbPool.Begin(nCtx)
		if err != nil {
			return err
		}
		defer tx.Rollback(nCtx)

		for _, metric := range metrics {
			switch metric.MType {
			case Gauge:
				if metric.Value == nil {
					return ErrMissingField
				}
				_, err = tx.Exec(nCtx, sqlAddMetricQuery, metric.ID, Gauge, metric.Value)
				if err != nil {
					return err
				}
			case Counter:
				if metric.Delta == nil {
					return ErrMissingField
				}

				counterMetrics[metric.ID] += *metric.Delta
			default:
				return ErrMetricNotSupportedType
			}
		}

		for metricID, aggregatedValue := range counterMetrics {
			pMetric, err := p.GetMetric(nCtx, Counter, metricID)
			if err != nil && !errors.Is(err, ErrMetricNotFound) {
				return err
			}

			if !errors.Is(err, ErrMetricNotFound) {
				aggregatedValue += *pMetric.Delta
			}

			_, err = tx.Exec(nCtx, sqlAddMetricQuery, metricID, Counter, strconv.Itoa(int(aggregatedValue)))
			if err != nil {
				return err
			}
		}

		err = tx.Commit(nCtx)
		return err
	})

	return err

}
