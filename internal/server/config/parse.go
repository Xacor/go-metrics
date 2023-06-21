package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "server address")
	flag.StringVar(&c.LogLevel, "l", "info", "log level")
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
