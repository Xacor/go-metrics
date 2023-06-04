package storage

import (
	"errors"
	"log"
	"sync"

	"github.com/Xacor/go-metrics/internal/model"
)

type MetricRepo interface {
	Get(id string) (model.Metric, error)
	Create(model.Metric) (model.Metric, error)
	Update(model.Metric) (model.Metric, error)
}

type MemStorage struct {
	data map[string]model.Metric
	mu   sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]model.Metric),
	}
}

func (mem *MemStorage) Get(id string) (model.Metric, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	val, ok := mem.data[id]
	if !ok {
		return model.Metric{}, errors.New("metric with this id not found")
	}
	return val, nil
}

func (mem *MemStorage) Create(metric model.Metric) (model.Metric, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	log.Println("create")
	_, exist := mem.data[metric.ID]
	if exist {
		return model.Metric{}, errors.New("metric with this id already exists")
	}

	mem.data[metric.ID] = metric
	return mem.data[metric.ID], nil
}

func (mem *MemStorage) Update(metric model.Metric) (model.Metric, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()

	// получение существующего экземпляра
	obj, exist := mem.data[metric.ID]
	if !exist {
		return model.Metric{}, errors.New("metric doesnt exist")
	}
	log.Println(obj)

	// изменение в зависимости от типа
	var err error
	switch obj.Type {
	case model.Counter:
		log.Println("counter type add")
		err = obj.Add(metric.Value)

	case model.Guage:
		log.Println("gauge type set")
		err = obj.Set(metric.Value)
	}

	if err != nil {
		return model.Metric{}, err
	}

	// запись в мапу
	mem.data[metric.ID] = obj

	return mem.data[metric.ID], nil
}
