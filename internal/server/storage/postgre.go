package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PostgreStorage struct {
	db *sql.DB
	l  *zap.Logger
}

func NewPostgreStorage(ctx context.Context, dsn string, logger *zap.Logger) (*PostgreStorage, error) {
	if dsn == "" {
		return nil, ErrEmptyDSN
	}
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.PingContext(ctx); err != nil {
		return nil, err
	}

	postgre := PostgreStorage{db: conn, l: logger}
	if err := postgre.Migrate(ctx); err != nil {
		return nil, err
	}

	return &postgre, nil
}

func (s *PostgreStorage) Migrate(ctx context.Context) error {
	var pgerr *pgconn.PgError

	createType := `CREATE TABLE IF NOT EXISTS metric_types (
        type VARCHAR(10) PRIMARY KEY
	);`
	_, err := s.db.ExecContext(ctx, createType)
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			err = s.retryExecContext(ctx, createType)
		}
		if err != nil {
			return err
		}
	}

	insertType := `INSERT INTO metric_types (type) VALUES 
		('counter'),
		('gauge')
		ON CONFLICT DO NOTHING
		;`
	_, err = s.db.ExecContext(ctx, insertType)
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			err = s.retryExecContext(ctx, insertType)
		}
		if err != nil {
			return err
		}
	}

	createMetrics := `CREATE TABLE IF NOT EXISTS metrics (
    	id serial PRIMARY KEY, 
        name VARCHAR(255) UNIQUE NOT NULL,
		mtype VARCHAR(10) references metric_types(type) NOT NULL,
        delta BIGINT CHECK ((delta IS NOT NULL AND mtype = 'counter') OR (delta IS NULL AND mtype = 'gauge')),
        value DOUBLE PRECISION CHECK ((value IS NOT NULL AND mtype = 'gauge') OR (value IS NULL AND mtype = 'counter'))
	);`
	_, err = s.db.ExecContext(ctx, createMetrics)
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			err = s.retryExecContext(ctx, createMetrics)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgreStorage) All(ctx context.Context) ([]model.Metrics, error) {
	var pgerr *pgconn.PgError

	query := "SELECT name, mtype, delta, value FROM metrics;"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			rows, err = s.retryQueryContext(ctx, query)
		}
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var metrics []model.Metrics

	for rows.Next() {
		var (
			name  string
			mtype string
			delta sql.NullInt64
			value sql.NullFloat64
		)

		if err := rows.Scan(&name, &mtype, &delta, &value); err != nil {
			return metrics, err
		}

		var m model.Metrics
		if delta.Valid {
			m = model.Metrics{
				Name:  name,
				MType: mtype,
				Delta: &delta.Int64,
			}
		} else if value.Valid {
			m = model.Metrics{
				Name:  name,
				MType: mtype,
				Value: &value.Float64,
			}
		} else {
			return metrics, ErrInvalidMetric
		}
		metrics = append(metrics, m)
	}

	if err = rows.Err(); err != nil {
		return metrics, err
	}
	return metrics, nil
}

func (s *PostgreStorage) Get(ctx context.Context, name string) (model.Metrics, error) {
	var pgerr *pgconn.PgError
	var m model.Metrics

	query := "SELECT name, mtype, delta, value FROM metrics WHERE name = $1;"

	row := s.db.QueryRowContext(ctx, query, name)
	err := row.Err()
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			row = s.retryQueryRowContext(ctx, query)
			err = row.Err()
		}
		if err != nil {
			return m, err
		}
	}

	var (
		mname string
		mtype string
		delta sql.NullInt64
		value sql.NullFloat64
	)

	if err := row.Scan(&mname, &mtype, &delta, &value); err != nil {
		return m, err
	}

	if delta.Valid {
		m = model.Metrics{
			Name:  mname,
			MType: mtype,
			Delta: &delta.Int64,
		}
	} else if value.Valid {
		m = model.Metrics{
			Name:  mname,
			MType: mtype,
			Value: &value.Float64,
		}
	} else {
		return m, ErrInvalidMetric
	}
	return m, nil
}

func (s *PostgreStorage) Create(ctx context.Context, m model.Metrics) (model.Metrics, error) {
	var pgerr *pgconn.PgError
	insert := "INSERT INTO metrics (name, mtype, delta, value) VALUES($1,$2,$3,$4);"

	_, err := s.db.ExecContext(ctx, insert, m.Name, m.MType, m.Delta, m.Value)
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			err = s.retryExecContext(ctx, insert, m.Name, m.MType, m.Delta, m.Value)
		}
		if err != nil {
			return model.Metrics{}, err
		}
	}

	return s.Get(ctx, m.Name)
}

func (s *PostgreStorage) Update(ctx context.Context, m model.Metrics) (model.Metrics, error) {
	var pgerr *pgconn.PgError
	update := "UPDATE metrics SET delta = metrics.delta + $1, value = $2 WHERE name = $3;"

	_, err := s.db.ExecContext(ctx, update, m.Delta, m.Value, m.Name)
	if err != nil {
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			err = s.retryExecContext(ctx, update, m.Delta, m.Value, m.Name)
		}
		if err != nil {
			return model.Metrics{}, err
		}
	}

	return s.Get(ctx, m.Name)
}

func (s *PostgreStorage) UpdateBatch(ctx context.Context, metrics []model.Metrics) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	upsert, err := tx.PrepareContext(ctx,
		`INSERT INTO metrics (name, mtype, delta, value) 
		VALUES($1,$2,$3,$4) 
		ON CONFLICT ON CONSTRAINT metrics_name_key 
		DO
		UPDATE SET delta = metrics.delta + $3, value = $4;`,
	)
	if err != nil {
		return err
	}
	defer upsert.Close()

	for _, m := range metrics {
		if _, err := upsert.ExecContext(ctx, m.Name, m.MType, m.Delta, m.Value); err != nil {
			var pgerr *pgconn.PgError

			if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
				err = s.retryExecPrepareContext(ctx, upsert, m.Name, m.MType, m.Delta, m.Value)
			}
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *PostgreStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgreStorage) Close() error {
	return s.db.Close()
}

func (s *PostgreStorage) retryExecContext(ctx context.Context, query string, args ...any) error {
	attempts := 0
	var err error
	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if _, err = s.db.ExecContext(ctx, query, args); err == nil {
			return nil
		}
		attempts++
		s.l.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
	}
	return err
}

func (s *PostgreStorage) retryQueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	attempts := 0

	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if rows, err = s.db.QueryContext(ctx, query, args); err == nil {
			return rows, err
		}
		attempts++
		s.l.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
	}
	return rows, err
}

func (s *PostgreStorage) retryQueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	var row *sql.Row

	attempts := 0

	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if row = s.db.QueryRowContext(ctx, query, args); row.Err() == nil {
			return row
		}
		attempts++
		s.l.Error("attempt failed", zap.Error(row.Err()), zap.Int("attempt #", attempts))
	}
	return row
}

func (s *PostgreStorage) retryExecPrepareContext(ctx context.Context, stmt *sql.Stmt, args ...any) error {
	attempts := 0
	var err error
	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if _, err = stmt.ExecContext(ctx, args); err == nil {
			return nil
		}
		attempts++
		s.l.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
	}
	return err
}
