package config

import (
	"fmt"
	"os"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
	Key            string `env:"KEY"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
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
