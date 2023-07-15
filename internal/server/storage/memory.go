package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/Xacor/go-metrics/internal/server/model"
)

type MemStorage struct {
	data map[string]model.Metrics
	mu   sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]model.Metrics),
	}
}

func (mem *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (mem *MemStorage) All(ctx context.Context) ([]model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	result := make([]model.Metrics, 0, len(mem.data))
	for _, v := range mem.data {
		result = append(result, v)
	}

	return result, nil
}

func (mem *MemStorage) Get(ctx context.Context, name string) (model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	val, ok := mem.data[name]
	if !ok {
		return model.Metrics{}, errors.New("metric with this Name not found")
	}
	return val, nil
}

func (mem *MemStorage) Create(ctx context.Context, metric model.Metrics) (model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	_, exist := mem.data[metric.Name]
	if exist {
		return model.Metrics{}, errors.New("metric with this Name already exists")
	}

	mem.data[metric.Name] = metric
	return mem.data[metric.Name], nil
}

func (mem *MemStorage) Update(ctx context.Context, metric model.Metrics) (model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	// получение существующего экземпляра
	obj, exist := mem.data[metric.Name]
	if !exist {
		return model.Metrics{}, errors.New("metric doesnt exist")
	}

	// изменение в зависимости от типа
	switch obj.MType {
	case model.TypeCounter:
		addDelta(metric.Delta, &obj)

	case model.TypeGauge:
		setValue(metric.Value, &obj)
	}

	// запись в мапу
	mem.data[metric.Name] = obj

	return mem.data[metric.Name], nil
}

func (mem *MemStorage) UpdateBatch(ctx context.Context, metrics []model.Metrics) error {
	for _, m := range metrics {
		_, exist := mem.Get(ctx, m.Name)
		if exist == nil {
			if _, err := mem.Create(ctx, m); err != nil {
				return err
			}
		} else {
			if _, err := mem.Update(ctx, m); err != nil {
				return err
			}
		}
	}

	return nil
}

func addDelta(delta *int64, dst *model.Metrics) {
	*dst.Delta = *dst.Delta + *delta
}

func setValue(value *float64, dst *model.Metrics) {
	*dst.Value = *value
}
