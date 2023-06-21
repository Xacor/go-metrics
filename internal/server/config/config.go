package config

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

func (c *Config) GetAddress() string {
	return c.Address
}
