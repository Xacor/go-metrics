package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "destination server address")
	flag.StringVar(&c.LogLevel, "l", "info", "log level")
	flag.IntVar(&c.ReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&c.PollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()
}

func (c *Config) ParseEnvs() error {
	err := env.Parse(c)

	return err
}

func (c *Config) ParseAll() error {
	c.ParseFlags()
	err := c.ParseEnvs()

	return err
}
