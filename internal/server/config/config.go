package config

import "os"

type Config struct {
	Address         string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	KeyFile         string `env:"KEY"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	Restore         bool   `env:"RESTORE"`
}

func (c *Config) GetKey() (string, error) {
	key, err := os.ReadFile(c.KeyFile)
	if err != nil {
		return "", err
	}

	return string(key), nil
}
