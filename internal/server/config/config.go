package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

type Config struct {
	Address              string `env:"ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
	FileStoragePath      string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN          string `env:"DATABASE_DSN"`
	KeyFile              string `env:"KEY"`
	CryptoKeyPrivateFile string `env:"CRYPTO_KEY"`
	StoreInterval        int    `env:"STORE_INTERVAL"`
	Restore              bool   `env:"RESTORE"`
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
