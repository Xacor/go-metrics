// Модуль storage содержит интерфейсы и их реализации для работы с БД.
package storage

import (
	"context"

	"github.com/Xacor/go-metrics/internal/server/model"
)

type Storage interface {
	MetricRepo
	Pinger
}

// Интерфейс хранилища метрик.
type MetricRepo interface {
	// All() возвращет список всех значений всех метрик.
	All(ctx context.Context) ([]model.Metrics, error)

	// Get() возвращет значение метрики по имени.
	Get(ctx context.Context, name string) (model.Metrics, error)

	// Create() сохраняет новую метрику.
	Create(ctx context.Context, metric model.Metrics) (model.Metrics, error)

	// Update() обновляет значение уже существующей метрки.
	Update(ctx context.Context, metric model.Metrics) (model.Metrics, error)

	// UpdateBatch() обновляет или, если неоходимо, создает метрики пачками.
	UpdateBatch(ctx context.Context, metrics []model.Metrics) error

	// Close() закрывает подключение к БД.
	Close() error
}

// Интерфейс позволяет проверить подключение к БД.
type Pinger interface {
	Ping(ctx context.Context) error
}
