package config

type Config struct {
	Address string `env:"ADDRESS"`
}

func (c *Config) GetAddress() string {
	return c.Address
}
