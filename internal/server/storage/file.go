package storage

import (
	"encoding/json"
	"os"

	"github.com/Xacor/go-metrics/internal/server/model"
)

func Save(path string, repo MetricRepo) error {
	data, err := repo.All()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(&data)

	return err
}

func Load(path string, repo MetricRepo) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var m model.Metrics

		err := decoder.Decode(&m)
		if err != nil {
			return err
		}

		_, err = repo.Create(m)
		if err != nil {
			return err
		}
	}

	return nil
}
