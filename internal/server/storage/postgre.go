package storage

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreStorage struct {
	db *sql.DB
}

func NewPostgreStorage(dsn string) (*PostgreStorage, error) {
	if dsn == "" {
		return nil, errors.New("empty data source")
	}
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &PostgreStorage{db: conn}, nil
}

func (s *PostgreStorage) Ping() error {
	return s.db.Ping()
}

func (s *PostgreStorage) Close() error {
	return s.db.Close()
}
