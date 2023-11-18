package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "server address")
	flag.StringVar(&c.GAddress, "g", "localhost:8081", "grpc server address")
	flag.StringVar(&c.GRPCConfig.TLSCertFile, "tls-cert", "", "tls cert file")
	flag.StringVar(&c.GRPCConfig.TLSKeyFile, "tls-key", "", "tls key file")
	flag.StringVar(&c.LogLevel, "l", "info", "log level")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database dsn e.g. host=127.0.0.1 port=5432 user=user dbname=db password=pass")
	flag.StringVar(&c.KeyFile, "k", "", "signature key")
	flag.StringVar(&c.CryptoKeyPrivateFile, "crypto-key", "", "path to RSA private key file in PEM format")
	flag.StringVar(&c.ConfigFile, "c", "", "path to configuration file")
	flag.StringVar(&c.TrustedSubnet, "t", "", "trusted subnet")
	flag.BoolVar(&c.Restore, "r", true, "leave true to restore previous state")
	flag.IntVar(&c.StoreInterval, "i", 300, "time between state saves")
	flag.Parse()
}

func (c *Config) ParseEnvs() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("failed to parse envs: %w", err)
	}

	return nil
}

func (c *Config) ParseFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) ParseAll() error {
	c.ParseFlags()

	if err := c.ParseEnvs(); err != nil {
		return err
	}

	if c.ConfigFile != "" {
		if err := c.ParseFile(c.ConfigFile); err != nil {
			return err
		}

		c.ParseFlags()

		if err := c.ParseEnvs(); err != nil {
			return err
		}
	}

	return nil
}
