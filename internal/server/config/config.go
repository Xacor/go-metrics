package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Address              string `env:"ADDRESS" json:"address"`
	LogLevel             string `env:"LOG_LEVEL" json:"log_level"`
	FileStoragePath      string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN          string `env:"DATABASE_DSN" json:"database_dsn"`
	KeyFile              string `env:"KEY" json:"key_file"`
	CryptoKeyPrivateFile string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigFile           string `env:"CONFIG" json:"-"`
	StoreInterval        int    `env:"STORE_INTERVAL" json:"store_interval"`
	Restore              bool   `env:"RESTORE" json:"restore"`
}

func (c *Config) GetKey() (string, error) {
	key, err := os.ReadFile(c.KeyFile)
	if err != nil {
		return "", err
	}

	return string(key), nil
}

func (c *Config) GetPrivateKey() (*rsa.PrivateKey, error) {
	key, err := os.ReadFile(c.CryptoKeyPrivateFile)
	if err != nil {
		fmt.Println("filename: ", c.CryptoKeyPrivateFile)
		return nil, err
	}

	rsaKey, err := bytesToPrivateKey(key)
	if err != nil {
		return nil, err
	}

	return rsaKey, nil
}

func bytesToPrivateKey(b []byte) (*rsa.PrivateKey, error) {
	var err error

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("unable to decode PEM private key")
	}

	blockBytes := block.Bytes
	private, err := x509.ParsePKCS1PrivateKey(blockBytes)
	if err != nil {
		return nil, err
	}

	return private, nil
}
