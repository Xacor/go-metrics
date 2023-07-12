package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreStorage struct {
	db         *sql.DB
	MetricRepo // !!!
}

func NewPostgreStorage(dsn string) (*PostgreStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// CHANGE INTERFACE INITIALIZATION!!!
	return &PostgreStorage{db: db, MetricRepo: NewMemStorage()}, nil
}

func (s *PostgreStorage) Ping() error {
	return s.db.Ping()
}

func (s *PostgreStorage) Close() error {
	return s.db.Close()
}
