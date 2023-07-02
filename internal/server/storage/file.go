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

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	json, err := json.MarshalIndent(data, "", "    ")
	json = append(json, '\n')

	_, err = file.Write(json)

	return err
}

func Load(path string, repo MetricRepo) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	_, err = decoder.Token()
	if err != nil {
		return err
	}
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

	_, err = decoder.Token()
	if err != nil {
		return err
	}

	return nil
}
