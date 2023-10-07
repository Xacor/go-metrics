package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Xacor/go-metrics/internal/server/model"
)

// Реализует логику для сохранения и загрузки метрик из файла.
type FileStorage struct {
	file *os.File
}

func NewFileStorage(path string) (*FileStorage, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("file %s not open: %w", path, err)
	}

	return &FileStorage{file: f}, nil
}

func (fs *FileStorage) Save(repo Storage) error {
	data, err := repo.All(context.Background())
	if err != nil {
		return err
	}
	fs.file.Truncate(0)
	fs.file.Seek(0, 0)

	enc := json.NewEncoder(fs.file)

	if err := enc.Encode(data); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) Load(repo Storage) error {

	decoder := json.NewDecoder(fs.file)

	m := make([]model.Metrics, 20)

	if err := decoder.Decode(&m); err != nil {
		return err
	}

	for _, v := range m {
		if _, err := repo.Create(context.Background(), v); err != nil {
			return err
		}
	}

	return nil
}
