package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Xacor/go-metrics/internal/server/model"
	"go.uber.org/zap"
)

type MemStorage struct {
	data          map[string]model.Metrics
	mu            sync.RWMutex
	fs            *FileStorage
	l             *zap.Logger
	storeInterval int
}

func NewMemStorage(backup *FileStorage, storeInterval int, logger *zap.Logger) *MemStorage {
	mem := &MemStorage{
		data:          make(map[string]model.Metrics),
		storeInterval: storeInterval,
		fs:            backup,
		l:             logger,
	}

	go mem.store()
	return mem
}

func (mem *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (mem *MemStorage) All(ctx context.Context) ([]model.Metrics, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	result := make([]model.Metrics, 0, len(mem.data))
	for _, v := range mem.data {
		result = append(result, v)
	}

	return result, nil
}

func (mem *MemStorage) Get(ctx context.Context, name string) (model.Metrics, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()

	val, ok := mem.data[name]
	if !ok {
		return model.Metrics{}, fmt.Errorf("%w: %s", ErrMetricNotFound, name)
	}
	return val, nil
}

func (mem *MemStorage) Create(ctx context.Context, metric model.Metrics) (model.Metrics, error) {
	_, err := mem.Get(ctx, metric.Name)
	if !errors.Is(err, ErrMetricNotFound) {
		return model.Metrics{}, ErrMetricExists
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	mem.data[metric.Name] = metric

	if err := mem.syncStore(); err != nil {
		mem.l.Error("failed to syncStore", zap.Error(err))
	}

	return mem.data[metric.Name], nil
}

func (mem *MemStorage) Update(ctx context.Context, metric model.Metrics) (model.Metrics, error) {

	obj, err := mem.Get(ctx, metric.Name)
	if errors.Is(err, ErrMetricNotFound) {
		return model.Metrics{}, err
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	switch obj.MType {
	case model.TypeCounter:
		addDelta(metric.Delta, &obj)

	case model.TypeGauge:
		setValue(metric.Value, &obj)
	}

	mem.data[metric.Name] = obj

	if err := mem.syncStore(); err != nil {
		mem.l.Error("failed to syncStore", zap.Error(err))
	}

	return mem.data[metric.Name], nil
}

func (mem *MemStorage) UpdateBatch(ctx context.Context, metrics []model.Metrics) error {
	for _, m := range metrics {
		_, ok := mem.Get(ctx, m.Name)
		if ok != nil {
			if _, err := mem.Create(ctx, m); err != nil {
				return fmt.Errorf("%w: %v", ErrMetricNotCreated, m)
			}
		} else {
			if _, err := mem.Update(ctx, m); err != nil {
				return fmt.Errorf("%w: %v", ErrMetricNotUpdated, m)
			}
		}
	}

	return nil
}

func (mem *MemStorage) Close() error {
	if err := mem.fs.Save(mem); err != nil {
		return err
	}

	return nil
}

func (mem *MemStorage) store() {
	if mem.storeInterval <= 0 {
		return
	}

	t := time.NewTicker(time.Duration(mem.storeInterval) * time.Second)
	for range t.C {
		mem.l.Debug("saving current state")
		if err := mem.fs.Save(mem); err != nil {
			mem.l.Error("failed to save data to file", zap.Error(err))
		}
	}
}

func (mem *MemStorage) syncStore() error {
	if mem.storeInterval != 0 {
		return nil
	}

	return mem.fs.Save(mem)
}

func addDelta(delta *int64, dst *model.Metrics) {
	*dst.Delta = *dst.Delta + *delta
}

func setValue(value *float64, dst *model.Metrics) {
	*dst.Value = *value
}
