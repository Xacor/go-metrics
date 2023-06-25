package storage

import (
	"errors"
	"sync"

	"github.com/Xacor/go-metrics/internal/server/model"
)

type MetricRepo interface {
	All() ([]model.Metrics, error)
	Get(id string) (model.Metrics, error)
	Create(model.Metrics) (model.Metrics, error)
	Update(model.Metrics) (model.Metrics, error)
}

type MemStorage struct {
	data map[string]model.Metrics
	mu   sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]model.Metrics),
	}
}

func (mem *MemStorage) All() ([]model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	result := make([]model.Metrics, 0, len(mem.data))
	for _, v := range mem.data {
		result = append(result, v)
	}

	return result, nil
}

func (mem *MemStorage) Get(id string) (model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	val, ok := mem.data[id]
	if !ok {
		return model.Metrics{}, errors.New("metric with this id not found")
	}
	return val, nil
}

func (mem *MemStorage) Create(metric model.Metrics) (model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	_, exist := mem.data[metric.ID]
	if exist {
		return model.Metrics{}, errors.New("metric with this id already exists")
	}

	mem.data[metric.ID] = metric
	return mem.data[metric.ID], nil
}

func (mem *MemStorage) Update(metric model.Metrics) (model.Metrics, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	// получение существующего экземпляра
	obj, exist := mem.data[metric.ID]
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
	mem.data[metric.ID] = obj

	return mem.data[metric.ID], nil
}

func addDelta(delta *int64, dst *model.Metrics) {
	*dst.Delta = *dst.Delta + *delta
}

func setValue(value *float64, dst *model.Metrics) {
	*dst.Value = *value
}
