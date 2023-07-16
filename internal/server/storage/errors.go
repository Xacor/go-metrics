package storage

import "errors"

var (
	ErrMetricNotFound   = errors.New("metric not found")
	ErrMetricExists     = errors.New("metric already exists")
	ErrMetricNotCreated = errors.New("failed to create metric")
	ErrMetricNotUpdated = errors.New("failed to update metric")
	ErrEmptyDSN         = errors.New("empty dsn")
	ErrTableCreation    = errors.New("failed to create table")
	ErrMigrationFailed  = errors.New("migration failed")
	ErrInvalidMetric    = errors.New("invalid metric values")
)
