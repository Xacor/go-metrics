package storage

import (
	"context"
	"database/sql"

	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PostgreStorage struct {
	db *pgxpool.Pool
	l  *zap.Logger
}

type sqlResponse struct {
	name  string
	mtype string
	delta sql.NullInt64
	value sql.NullFloat64
}

func NewPostgreStorage(ctx context.Context, dsn string, logger *zap.Logger) (*PostgreStorage, error) {
	if dsn == "" {
		return nil, ErrEmptyDSN
	}
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	postgre := PostgreStorage{db: conn, l: logger}
	if err := postgre.Migrate(ctx); err != nil {
		return nil, err
	}

	return &postgre, nil
}

func (s *PostgreStorage) Migrate(ctx context.Context) error {
	// var pgerr *pgconn.PgError

	createType := `CREATE TABLE IF NOT EXISTS metric_types (
        type VARCHAR(10) PRIMARY KEY
	);`
	_, err := s.db.Exec(ctx, createType)
	if err != nil {
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	err = s.retryExecContext(ctx, createType)
		// }
		if err != nil {
			return err
		}
	}

	insertType := `INSERT INTO metric_types (type) VALUES 
		('counter'),
		('gauge')
		ON CONFLICT DO NOTHING
		;`
	_, err = s.db.Exec(ctx, insertType)
	if err != nil {
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	err = s.retryExecContext(ctx, insertType)
		// }
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
	_, err = s.db.Exec(ctx, createMetrics)
	if err != nil {
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	err = s.retryExecContext(ctx, createMetrics)
		// }
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgreStorage) All(ctx context.Context) ([]model.Metrics, error) {

	query := "SELECT name, mtype, delta, value FROM metrics;"
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		// var pgerr *pgconn.PgError
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	rows, err = s.retryQueryContext(ctx, query)
		// }
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var metrics []model.Metrics

	for rows.Next() {
		var sql sqlResponse

		if err := rows.Scan(&sql.name, &sql.mtype, &sql.delta, &sql.value); err != nil {
			return metrics, err
		}

		m := model.Metrics{Name: sql.name, MType: sql.mtype}

		if sql.delta.Valid {
			m.Delta = &sql.delta.Int64

		} else if sql.value.Valid {

			m.Value = &sql.value.Float64
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
	query := "SELECT name, mtype, delta, value FROM metrics WHERE name = $1;"

	var sql sqlResponse

	err := s.db.QueryRow(ctx, query, name).Scan(&sql.name, &sql.mtype, &sql.delta, &sql.value)
	if err != nil {
		// var pgerr *pgconn.PgError
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	row := s.retryQueryRowContext(ctx, query)
		// 	err = row.Err()
		// }
		if err != nil {
			return model.Metrics{}, err
		}
	}

	m := model.Metrics{Name: sql.name, MType: sql.mtype}

	if sql.delta.Valid {
		m.Delta = &sql.delta.Int64

	} else if sql.value.Valid {

		m.Value = &sql.value.Float64
	} else {
		return model.Metrics{}, ErrInvalidMetric
	}
	return m, nil
}

func (s *PostgreStorage) Create(ctx context.Context, m model.Metrics) (model.Metrics, error) {

	insert := "INSERT INTO metrics (name, mtype, delta, value) VALUES($1,$2,$3,$4);"

	_, err := s.db.Exec(ctx, insert, m.Name, m.MType, m.Delta, m.Value)
	if err != nil {
		// var pgerr *pgconn.PgError
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	err = s.retryExecContext(ctx, insert, m.Name, m.MType, m.Delta, m.Value)
		// }
		if err != nil {
			return model.Metrics{}, err
		}
	}

	return s.Get(ctx, m.Name)
}

func (s *PostgreStorage) Update(ctx context.Context, m model.Metrics) (model.Metrics, error) {

	update := "UPDATE metrics SET delta = metrics.delta + $1, value = $2 WHERE name = $3;"

	_, err := s.db.Exec(ctx, update, m.Delta, m.Value, m.Name)
	if err != nil {
		// var pgerr *pgconn.PgError
		// if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
		// 	err = s.retryExecContext(ctx, update, m.Delta, m.Value, m.Name)
		// }
		if err != nil {
			return model.Metrics{}, err
		}
	}

	return s.Get(ctx, m.Name)
}

func (s *PostgreStorage) UpdateBatch(ctx context.Context, metrics []model.Metrics) error {
	batch := &pgx.Batch{}
	query := `INSERT INTO metrics (name, mtype, delta, value) 
		VALUES($1,$2,$3,$4) 
		ON CONFLICT ON CONSTRAINT metrics_name_key 
		DO
		UPDATE SET delta = metrics.delta + $3, value = $4;`

	for _, m := range metrics {
		batch.Queue(query, m.Name, m.MType, m.Delta, m.Value)
	}

	return s.db.SendBatch(ctx, batch).Close()
}

func (s *PostgreStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *PostgreStorage) Close() error {
	s.db.Close()

	return nil
}

// func (s *PostgreStorage) retryExecContext(ctx context.Context, query string, args ...any) error {
// 	attempts := 0
// 	var err error
// 	for i := 1; i < 5; i += 2 {
// 		time.Sleep(time.Second * time.Duration(i))
// 		if _, err = s.db.Exec(ctx, query, args); err == nil {
// 			return nil
// 		}
// 		attempts++
// 		s.l.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
// 	}
// 	return err
// }

// func (s *PostgreStorage) retryQueryContext(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
// 	var rows pgx.Rows
// 	var err error

// 	attempts := 0

// 	for i := 1; i < 5; i += 2 {
// 		time.Sleep(time.Second * time.Duration(i))
// 		if rows, err = s.db.Query(ctx, query, args); err == nil {
// 			return rows, err
// 		}
// 		attempts++
// 		s.l.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
// 	}
// 	return rows, err
// }

// func (s *PostgreStorage) retryQueryRowContext(ctx context.Context, query string, args ...any) pgx.Row {
// 	var row pgx.Row

// 	attempts := 0

// 	for i := 1; i < 5; i += 2 {
// 		time.Sleep(time.Second * time.Duration(i))
// 		if row = s.db.QueryRow(ctx, query, args); row.Err() == nil {
// 			return row
// 		}
// 		attempts++
// 		s.l.Error("attempt failed", zap.Error(row.Err()), zap.Int("attempt #", attempts))
// 	}
// 	return row
// }

// func (s *PostgreStorage) retryExecPrepareContext(ctx context.Context, stmt *sql.Stmt, args ...any) error {
// 	attempts := 0
// 	var err error
// 	for i := 1; i < 5; i += 2 {
// 		time.Sleep(time.Second * time.Duration(i))
// 		if _, err = stmt.ExecContext(ctx, args); err == nil {
// 			return nil
// 		}
// 		attempts++
// 		s.l.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
// 	}
// 	return err
// }
