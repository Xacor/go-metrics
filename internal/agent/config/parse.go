package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "destination server address")
	flag.StringVar(&c.LogLevel, "log", "info", "log level")
	flag.StringVar(&c.Key, "k", "", "signature key")
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

func (c *Config) ParseAll() error {
	c.ParseFlags()
	err := c.ParseEnvs()

	return err
}
