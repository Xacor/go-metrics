package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

func (c *Config) ParseFlags() {
	flag.StringVar(&c.Address, "a", "localhost:8080", "server address")
	flag.StringVar(&c.LogLevel, "l", "info", "log level")
	flag.StringVar(&c.FileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	flag.StringVar(&c.DatabaseDSN, "d", "", "database dsn e.g. host=127.0.0.1 port=5432 user=user dbname=db password=pass")
	flag.BoolVar(&c.Restore, "r", true, "leave true to restore previous state")
	flag.IntVar(&c.StoreInterval, "i", 300, "time between state saves")
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
