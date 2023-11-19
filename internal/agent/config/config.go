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
	GRPCConfig
	Address             string `env:"ADDRESS" json:"address"`
	LogLevel            string `env:"LOG_LEVEL" json:"log_level"`
	Key                 string `env:"KEY" json:"key"`
	CryptoKeyPublicFile string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigFile          string `env:"CONFIG" json:"-"`
	ReportInterval      int    `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval        int    `env:"POLL_INTERVAL" json:"poll_interval"`
	RateLimit           int    `env:"RATE_LIMIT" json:"rate_limit"`
}

type GRPCConfig struct {
	GRPCAddress string `env:"G_ADDRESS" json:"g_address"`
	CACertFile  string `env:"CA_CERT" json:"ca_cert"`
}

func (c *Config) GetURL() string {
	return fmt.Sprintf("http://%s", c.Address)
}

func (c *Config) GetReportInterval() int {
	return c.ReportInterval
}

func (c *Config) GetPollInterval() int {
	return c.PollInterval
}

func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

func (c *Config) GetKey() (string, error) {
	key, err := os.ReadFile(c.Key)
	if err != nil {
		return "", err
	}
	return string(key), nil
}

func (c *Config) GetRateLimit() int {
	return c.RateLimit
}

func (c *Config) GetPublicKey() (*rsa.PublicKey, error) {
	if c.CryptoKeyPublicFile == "" {
		return nil, nil
	}

	key, err := os.ReadFile(c.CryptoKeyPublicFile)
	if err != nil {
		return nil, err
	}

	rsaKey, err := bytesToPublicKey(key)
	if err != nil {
		return nil, err
	}

	return rsaKey, nil
}

func bytesToPublicKey(b []byte) (*rsa.PublicKey, error) {
	var err error

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("unable to decode PEM private key")
	}

	blockBytes := block.Bytes
	public, err := x509.ParsePKCS1PublicKey(blockBytes)
	if err != nil {
		return nil, err
	}

	return public, nil
}
