package storage

import (
	"encoding/json"
	"os"

	"github.com/Xacor/go-metrics/internal/server/model"
)

type FileStorage struct {
	file *os.File
}

func NewFileStorage(path string) (*FileStorage, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{file: f}, nil
}

func (fs *FileStorage) Save(repo MetricRepo) error {
	data, err := repo.All()
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

func (fs *FileStorage) Load(repo MetricRepo) error {

	decoder := json.NewDecoder(fs.file)

	m := make([]model.Metrics, 20)

	if err := decoder.Decode(&m); err != nil {
		return err
	}

	for _, v := range m {
		if _, err := repo.Create(v); err != nil {
			return err
		}
	}

	return nil
}
