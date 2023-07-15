package storage

import (
	"context"

	"github.com/Xacor/go-metrics/internal/server/model"
)

type Storage interface {
	MetricRepo
	Pinger
}

type MetricRepo interface {
	All(ctx context.Context) ([]model.Metrics, error)
	Get(ctx context.Context, name string) (model.Metrics, error)
	Create(ctx context.Context, metric model.Metrics) (model.Metrics, error)
	Update(ctx context.Context, metric model.Metrics) (model.Metrics, error)
	UpdateBatch(ctx context.Context, metrics []model.Metrics) error
}

type Pinger interface {
	Ping(ctx context.Context) error
}
