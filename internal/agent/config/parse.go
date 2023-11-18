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
	flag.StringVar(&c.Address, "a", "localhost:8080", "destination server address")
	flag.StringVar(&c.GRPCAddress, "g", "localhost:8081", "destination server address")
	flag.StringVar(&c.CACertFile, "ca", "", "ca cert file")
	flag.StringVar(&c.LogLevel, "log", "info", "log level")
	flag.StringVar(&c.Key, "k", "", "signature key")
	flag.StringVar(&c.CryptoKeyPublicFile, "crypto-key", "", "path to RSA public key file in PEM format")
	flag.StringVar(&c.ConfigFile, "c", "", "path to configuration file")
	flag.IntVar(&c.ReportInterval, "r", 5, "report interval in seconds")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&c.RateLimit, "l", 1, "rate limit")
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
